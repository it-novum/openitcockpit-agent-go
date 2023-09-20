package webserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

const (
	testBasicAuthUser     = "user"
	testBasicAuthPassword = "test"
	testBasicAuth         = "user:test"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestWebserverHandler(t *testing.T) {
	stateInput := make(chan []byte)
	w := &handler{
		StateInput: stateInput,
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)
	stateInput <- []byte(`{"test": "tata"}`)
	cancel()
	w.wg.Wait()
}

func TestWebserverHandlerState(t *testing.T) {
	stateInput := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := &handler{
		StateInput: stateInput,
		Configuration: &config.Configuration{
			BasicAuth: "",
		},
	}
	ts := httptest.NewServer(w.Handler())
	defer ts.Close()

	w.Start(ctx)

	testState := []byte(`{"test": "tata"}`)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.SetBasicAuth(testBasicAuthUser, testBasicAuthPassword)

	r, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if string(body) != "{}" {
		t.Fatal("body does not match")
	}

	stateInput <- testState

	r, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err = io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if string(body) != string(testState) {
		t.Fatal("body does not match")
	}

	w.Shutdown()
}

func TestWebserverHandlerAuthFailed(t *testing.T) {
	stateInput := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := &handler{
		StateInput: stateInput,
		Configuration: &config.Configuration{
			BasicAuth: testBasicAuth,
		},
	}
	w.Start(ctx)

	ts := httptest.NewServer(w.Handler())
	defer ts.Close()

	stateInput <- []byte(`{"test": "tata"}`)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.SetBasicAuth("some", "fake")

	r, err := client.Do(req)
	if err != nil {
		t.Fatal("Request failed")
	}
	if r.StatusCode != http.StatusForbidden {
		t.Error("Unexpected status code: ", http.StatusForbidden)
	}
	_ = r.Body.Close()

	req, _ = http.NewRequest("GET", ts.URL, nil)
	r, err = client.Do(req)
	if err != nil {
		t.Fatal("Request failed")
	}
	if r.StatusCode != http.StatusForbidden {
		t.Error("Unexpected status code: ", http.StatusForbidden)
	}
	_ = r.Body.Close()

	w.Shutdown()
}

func TestWebserverHandlerConfig(t *testing.T) {
	state := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpdir, err := os.MkdirTemp(os.TempDir(), "*-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpdir)
	}()
	cfgPath := filepath.Join(tmpdir, "config.ini")

	w := &handler{
		StateInput: state,
		Configuration: &config.Configuration{
			ConfigurationPath: cfgPath,
			ConfigUpdate:      true,
		},
	}

	ts := httptest.NewServer(w.Handler())
	defer ts.Close()

	result := ""
	w.Start(ctx)

	data, err := json.Marshal(&configurationPush{
		Configuration:            base64.StdEncoding.EncodeToString([]byte(`[default]`)),
		CustomCheckConfiguration: "",
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(ts.URL+"/config", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Status code is not 200")
	}

	d, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Error(err)
	} else {
		result = string(d)
		if result != `[default]` {
			t.Error("unexpected result ([default]): ", result)
		}
	}

	resp, err = http.Get(ts.URL + "/config")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Status code is not 200")
	}

	cp := &configurationPush{}
	if err := json.Unmarshal(body, cp); err != nil {
		t.Fatal(err)
	}
	cfg, err := base64.StdEncoding.DecodeString(cp.Configuration)
	if err != nil {
		t.Fatal(err)
	}
	if string(cfg) != "[default]" {
		t.Fatal("unexpected response for configuration get")
	}

	w.Shutdown()
}

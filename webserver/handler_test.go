package webserver

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

const (
	testBasicAuthUser     = "user"
	testBasicAuthPassword = "test"
	testBasicAuth         = "user:test"
)

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
	body, err := ioutil.ReadAll(r.Body)
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
	body, err = ioutil.ReadAll(r.Body)
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
		t.Error("Request failed")
	}
	if r.StatusCode != http.StatusForbidden {
		t.Error("Unexpected status code: ", http.StatusForbidden)
	}
	r.Body.Close()

	req, _ = http.NewRequest("GET", ts.URL, nil)
	r, err = client.Do(req)
	if err != nil {
		t.Error("Request failed")
	}
	if r.StatusCode != http.StatusForbidden {
		t.Error("Unexpected status code: ", http.StatusForbidden)
	}
	r.Body.Close()

	w.Shutdown()
}

func TestWebserverHandlerConfig(t *testing.T) {
	state := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpdir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	cfgPath := path.Join(tmpdir, "config.ini")

	w := &handler{
		StateInput: state,
		Configuration: &config.Configuration{
			ConfigurationPath: cfgPath,
		},
	}

	ts := httptest.NewServer(w.Handler())
	defer ts.Close()

	result := ""
	w.Start(ctx)

	cfg := "someconfig"
	resp, err := http.Post(ts.URL+"/config", "text/plain", strings.NewReader(cfg))
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Status code is not 200")
	}

	d, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		t.Error(err)
	} else {
		result = string(d)
		if result != cfg {
			t.Error("unexpected result")
		}
	}

	w.Shutdown()
}

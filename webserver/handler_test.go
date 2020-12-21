package webserver

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

var (
	testBasicAuth = &config.BasicAuth{
		Username: "test",
		Password: "test",
	}
)

func TestWebserverHandler(t *testing.T) {
	stateInput := make(chan []byte)
	w := &handler{
		ConfigPushRecipient: make(chan string),
		StateInput:          stateInput,
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.prepare()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()
	stateInput <- []byte(`{"test": "tata"}`)
	cancel()
	wg.Wait()
}

func TestWebserverHandlerState(t *testing.T) {
	stateInput := make(chan []byte)
	configPush := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := &handler{
		ConfigPushRecipient: configPush,
		StateInput:          stateInput,
		BasicAuthConfig:     testBasicAuth,
	}
	w.prepare()

	ts := httptest.NewServer(w.Handler())
	defer ts.Close()

	go w.Run(ctx)

	testState := []byte(`{"test": "tata"}`)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.SetBasicAuth(testBasicAuth.Username, testBasicAuth.Password)

	r, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
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
	if string(body) != string(testState) {
		t.Fatal("body does not match")
	}

	w.Shutdown()
}

func TestWebserverHandlerAuthFailed(t *testing.T) {
	stateInput := make(chan []byte)
	configPush := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := &handler{
		ConfigPushRecipient: configPush,
		StateInput:          stateInput,
		BasicAuthConfig:     testBasicAuth,
	}
	w.prepare()
	go w.Run(ctx)

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
	configPush := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := &handler{
		ConfigPushRecipient: configPush,
		StateInput:          state,
	}
	w.prepare()

	ts := httptest.NewServer(w.Handler())
	defer ts.Close()

	result := ""
	done := make(chan struct{})

	go func() {
		result = <-configPush
		done <- struct{}{}
	}()

	go w.Run(ctx)

	cfg := "someconfig"
	resp, err := http.Post(ts.URL+"/config", "text/plain", strings.NewReader(cfg))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Status code is not 200")
	}

	select {
	case <-done:
		if result != cfg {
			t.Error("unexpected result")
		}
	case <-time.After(time.Second * 2):
		t.Error("Timeout")
	}

	w.Shutdown()
}

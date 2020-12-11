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
)

func TestAgentWebserver(t *testing.T) {
	state := make(chan []byte)
	configPush := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		RunAgentWebserver(ctx, state, configPush)
	}()
	state <- []byte(`{"test": "tata"}`)
	cancel()
	wg.Wait()
}

func TestAgentWebserverState(t *testing.T) {
	state := make(chan []byte)
	configPush := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	w := &webserverHandler{
		configPush: configPush,
		stateInput: state,
	}

	ts := httptest.NewServer(w.handler())
	defer ts.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.run(ctx)
	}()

	testState := []byte(`{"test": "tata"}`)
	r, err := http.Get(ts.URL)
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

	state <- testState

	r, err = http.Get(ts.URL)
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
	cancel()
	wg.Wait()
}

func TestAgentWebserverConfig(t *testing.T) {
	state := make(chan []byte)
	configPush := make(chan string)
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	w := &webserverHandler{
		configPush: configPush,
		stateInput: state,
	}

	ts := httptest.NewServer(w.handler())
	defer ts.Close()

	result := ""
	done := make(chan struct{})

	go func() {
		result = <-configPush
		done <- struct{}{}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.run(ctx)
	}()

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

	cancel()
	wg.Wait()
}

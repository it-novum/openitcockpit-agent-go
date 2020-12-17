package webserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// TODO SSL
// TODO auth
// TODO autossl

type handler struct {
	StateInput          <-chan []byte
	ConfigPushRecipient chan<- string

	mtx    sync.RWMutex
	state  []byte
	router *mux.Router
}

func (w *handler) getState() []byte {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	if w.state == nil {
		return []byte("{}")
	}
	return w.state
}

func (w *handler) setState(newState []byte) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.state = newState
}

func (w *handler) handleStatus(response http.ResponseWriter, request *http.Request) {
	response.Write(w.getState())
	response.WriteHeader(200)
}

func (w *handler) handleConfig(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)
		response.WriteHeader(500)
		response.Write([]byte("could not read body"))
	}
	w.ConfigPushRecipient <- string(body)
}

// Handler can be used by http.Server to handle http connections
func (w *handler) Handler() *mux.Router {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.router == nil {
		routes := mux.NewRouter()
		routes.HandleFunc("/", w.handleStatus)
		routes.HandleFunc("/config", w.handleConfig)
		w.router = routes
	}
	return w.router
}

// Run webserver handler
// (should be run in a go routine)
func (w *handler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-w.StateInput:
			w.setState(s)
		}
	}
}

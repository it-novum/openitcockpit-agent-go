package webserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// TODO SSL
// TODO auth
// TODO autossl

type webserverHandler struct {
	mtx        sync.RWMutex
	stateInput <-chan []byte
	state      []byte
	configPush chan<- string
}

func (w *webserverHandler) getState() []byte {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	if w.state == nil {
		return []byte("{}")
	}
	return w.state
}

func (w *webserverHandler) setState(newState []byte) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.state = newState
}

func (w *webserverHandler) handleStatus(response http.ResponseWriter, request *http.Request) {
	response.Write(w.getState())
	response.WriteHeader(200)
}

func (w *webserverHandler) handleConfig(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)
		response.WriteHeader(500)
		response.Write([]byte("internal server error"))
	}
	w.configPush <- string(body)
}

func (w *webserverHandler) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-w.stateInput:
			w.setState(s)
		}
	}
}

func (w *webserverHandler) handler() *mux.Router {
	routes := mux.NewRouter()
	routes.HandleFunc("/", w.handleStatus)
	routes.HandleFunc("/config", w.handleConfig)
	return routes
}

// RunAgentWebserver starts a http server handling all requests
// (should be run in a go routine)
func RunAgentWebserver(ctx context.Context, state <-chan []byte, configPush chan<- string) {
	w := &webserverHandler{
		configPush: configPush,
		stateInput: state,
	}

	server := &http.Server{
		Handler:        w.handler(),
		Addr:           ":3333",
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 30,
		IdleTimeout:    time.Second * 30,
		MaxHeaderBytes: 256 * 1024,
	}

	go server.ListenAndServe()
	w.run(ctx)

	sdctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// TODO log error
	server.Shutdown(sdctx)
}

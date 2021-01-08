package webserver

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	log "github.com/sirupsen/logrus"
)

type contextKey string

const authenticatedKey contextKey = "Authenticated"

type basicAuthMiddleware struct {
	Username string
	Password string
}

func (b *basicAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != b.Username || password != b.Password {
			log.Infoln("Webserver: Invalid username or password from client: ", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedKey, true)))
		}
	})
}

func tlsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedKey, true)))
		} else {
			log.Infoln("Webserver: No client certificate: ", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

type handler struct {
	StateInput <-chan []byte

	Configuration *config.Configuration

	mtx      sync.RWMutex
	shutdown chan struct{}
	state    []byte
	wg       sync.WaitGroup
	prepared bool

	router              *mux.Router
	basicAuthMiddleware *basicAuthMiddleware
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
	response.Header().Add("Content-Type", "application/json")
	response.Write(w.getState())
}

func (w *handler) handleConfigRead(response http.ResponseWriter, request *http.Request) {
}

func (w *handler) handleConfigPush(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Errorln("Webserver: Could not read body: ", err)
		http.Error(response, "could not read body", http.StatusInternalServerError)
		return
	}

	if err := config.SaveConfiguration(w.Configuration, body); err != nil {
		log.Errorln("Webserver: ", err)
	}
}

func (w *handler) handlerCsr(response http.ResponseWriter, request *http.Request) {
	if err := utils.GeneratePrivateKeyIfNotExists(w.Configuration.AutoSslKeyFile); err != nil {
		log.Errorln("Webserver: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}
	csr, err := utils.CSRFromKeyFile(w.Configuration.AutoSslKeyFile, request.URL.Query().Get("domain"))
	if err != nil {
		log.Errorln("Webserver: could not generate csr: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := ioutil.WriteFile(w.Configuration.AutoSslCsrFile, csr, 0666); err != nil {
		log.Infoln("Webserver: could not store csr: ", err)
	}
	js, err := json.Marshal(struct{ Csr string }{string(csr)})
	if err != nil {
		log.Errorln("Webserver: Could not create json for csr: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}
	response.Header().Add("Content-Type", "application/json")
	response.Write(js)
}

func (w *handler) handlerUpdateCert(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	log.Debugln("Webserver: Certificate update")

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Errorln("Webserver: Could not read body: ", err)
		http.Error(response, "could not read body", http.StatusInternalServerError)
		return
	}

	if err := ioutil.WriteFile(w.Configuration.AutoSslCrtFile, body, 0666); err != nil {
		log.Errorln("Webserver: Could not write certificate file: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Debugln("Webserver: Certificate update successful, start reload")
	w.Configuration.Reload()
}

// Handler can be used by http.Server to handle http connections
func (w *handler) Handler() *mux.Router {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.router == nil {
		routes := mux.NewRouter()
		if isAutosslEnabled(w.Configuration) {
			log.Infoln("Webserver: Activate TLS authentication")
			routes.Use(tlsAuthMiddleware)
		}
		if w.Configuration.BasicAuth != "" {
			log.Infoln("Webserver: Activate Basic authentication")
			cred := strings.SplitN(w.Configuration.BasicAuth, ":", 2)
			if len(cred) != 2 {
				log.Fatalln("Webserver: Invalid basic auth configuration")
			}
			w.basicAuthMiddleware = &basicAuthMiddleware{
				Username: cred[0],
				Password: cred[1],
			}
			routes.Use(w.basicAuthMiddleware.Middleware)
		}
		routes.Path("/").Methods("GET").HandlerFunc(w.handleStatus)
		routes.Path("/config").Methods("GET").HandlerFunc(w.handleConfigRead)
		// TODO disable if not need
		routes.Path("/config").Methods("POST").HandlerFunc(w.handleConfigPush)
		routes.Path("/getCsr").Methods("GET").HandlerFunc(w.handlerCsr)
		routes.Path("/updateCrt").Methods("POST").HandlerFunc(w.handlerUpdateCert)
		w.router = routes
	}
	return w.router
}

func (w *handler) Shutdown() {
	close(w.shutdown)
	w.wg.Wait()
}

// Start webserver handler (should NOT run in a go routine)
func (w *handler) Start(parentCtx context.Context) {
	w.shutdown = make(chan struct{})

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		ctx, cancel := context.WithCancel(parentCtx)
		defer cancel()

		log.Debugln("Webserver: Handler waiting for input")

		for {
			select {
			case _, more := <-w.shutdown:
				if !more {
					return
				}
			case <-ctx.Done():
				log.Debugln("Webserver: Handler ctx cancled")
				return
			case s := <-w.StateInput:
				w.setState(s)
			}
		}
	}()
}

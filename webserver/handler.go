package webserver

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	log "github.com/sirupsen/logrus"
)

type contextKey string

const authenticatedKey contextKey = "Authenticated"

type basicAuthMiddleware struct {
	BasicAuthConfig *config.BasicAuth
}

func (b *basicAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != b.BasicAuthConfig.Username || password != b.BasicAuthConfig.Password {
			log.Infoln("Webserver: Invalid username or password from client: ", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedKey, true)))
		}
	})
}

func tlsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TLS.PeerCertificates) > 0 {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedKey, true)))
		} else {
			log.Infoln("Webserver: No client certificate: ", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

type handler struct {
	StateInput          <-chan []byte
	ConfigPushRecipient chan<- string

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

func sendInternalServerError(response http.ResponseWriter, text string) {
	if text == "" {
		text = "internal server error"
	}
	response.Write([]byte(text))
	response.WriteHeader(500)
}

func (w *handler) handleStatus(response http.ResponseWriter, request *http.Request) {
	response.Write(w.getState())
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(200)
}

func (w *handler) handleConfigRead(response http.ResponseWriter, request *http.Request) {
}

func (w *handler) handleConfigPush(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Errorln("Webserver: Could not read body: ", err)
		sendInternalServerError(response, "could not read body")
		return
	}
	w.ConfigPushRecipient <- string(body)
}

func (w *handler) handlerCsr(response http.ResponseWriter, request *http.Request) {
	utils.GeneratePrivateKeyIfNotExists(w.Configuration.TLS.AutoSslKeyFile)
	csr, err := utils.CSRFromKeyFile(w.Configuration.TLS.AutoSslKeyFile, request.URL.Query().Get("domain"))
	if err != nil {
		log.Errorln("Webserver: could not generate csr: ", err)
		sendInternalServerError(response, "")
		return
	}
	if err := ioutil.WriteFile(w.Configuration.TLS.AutoSslCsrFile, csr, 0666); err != nil {
		log.Infoln("Webserver: could not store csr: ", err)
	}
	js, err := json.Marshal(struct{ Csr string }{string(csr)})
	if err != nil {
		log.Errorln("Webserver: Could not create json for csr: ", err)
		sendInternalServerError(response, "")
		return
	}
	response.Write(js)
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(200)
}

func (w *handler) handlerUpdateCert(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Errorln("Webserver: Could not read body: ", err)
		sendInternalServerError(response, "could not read body")
		return
	}
	if err := ioutil.WriteFile(w.Configuration.TLS.AutoSslCrtFile, body, 0666); err != nil {
		log.Errorln("Webserver: Could not write certificate file: ", err)
		sendInternalServerError(response, "")
		return
	}
}

// Handler can be used by http.Server to handle http connections
func (w *handler) Handler() *mux.Router {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.router == nil {
		routes := mux.NewRouter()
		if w.Configuration.TLS != nil && w.Configuration.TLS.AutoSslEnabled {
			log.Infoln("Webserver: Activate TLS authentication")
			routes.Use(tlsAuthMiddleware)
		}
		if w.Configuration.BasicAuth != nil && w.Configuration.BasicAuth.Username != "" {
			log.Infoln("Webserver: Activate Basic authentication")
			w.basicAuthMiddleware = &basicAuthMiddleware{
				BasicAuthConfig: w.Configuration.BasicAuth,
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

func (w *handler) prepare() {
	w.wg.Add(1)
	w.shutdown = make(chan struct{})
	w.prepared = true
}

// Run webserver handler
// (should be run in a go routine)
func (w *handler) Run(parentCtx context.Context) {
	if !w.prepared {
		log.Fatalln("Webserver: handler was not prepared")
	}

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
}

package webserver

import (
	"context"
	"encoding/base64"
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

func debugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugln("Webserver: Request: ", r.RemoteAddr, " ", r.Method, " ", r.URL)
		next.ServeHTTP(w, r)
	})
}

type csrResponse struct {
	Csr string `json:"csr"`
}

type updateCrtRequest struct {
	Signed string `json:"signed"`
	CA     string `json:"ca"`
}

type handler struct {
	StateInput    <-chan []byte
	Reloader      Reloader
	Configuration *config.Configuration

	mtx      sync.RWMutex
	shutdown chan struct{}
	state    []byte
	wg       sync.WaitGroup

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
	log.Debugln("Webserver: set new state")
	w.state = newState
}

func (w *handler) handleStatus(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_, err := response.Write(w.getState())
	if err != nil {
		log.Errorln("Webserver: ", err)
	}
}

type configurationPush struct {
	Configuration            string `json:"configuration"`
	CustomCheckConfiguration string `json:"customcheck_configuration"`
}

func (w *handler) handleConfigRead(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	if !w.Configuration.ConfigUpdate {
		http.Error(response, "config update is disabled", http.StatusForbidden)
		return
	}

	r := configurationPush{}

	data, err := w.Configuration.ReadConfigurationFile()
	if err != nil {
		log.Errorln("Webserver: Could not read configuration file: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}
	r.Configuration = base64.StdEncoding.EncodeToString(data)

	data = w.Configuration.ReadCustomCheckConfiguration()
	r.CustomCheckConfiguration = base64.StdEncoding.EncodeToString(data)

	data, err = json.Marshal(&r)
	if err != nil {
		log.Errorln("Webserver: Could not create json for configuration read: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}

	response.Header().Add("Content-Type", "application/json")
	if _, err := response.Write(data); err != nil {
		log.Errorln("Webserver: ", err)
	}
}

func (w *handler) handleConfigPush(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	if !w.Configuration.ConfigUpdate {
		http.Error(response, "config update is disabled", http.StatusForbidden)
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Errorln("Webserver: Could not read body: ", err)
		http.Error(response, "could not read body", http.StatusInternalServerError)
		return
	}

	r := configurationPush{}
	if err := json.Unmarshal(body, &r); err != nil {
		log.Errorln("Webserver: Could not parse json for configuration push: ", err)
		http.Error(response, "invalid json or base64 string", http.StatusInternalServerError)
		return
	}

	cfgData, err := base64.StdEncoding.DecodeString(r.Configuration)
	if err != nil {
		log.Errorln("Webserver: Could not decode configuration string for configuration push: ", err)
		http.Error(response, "invalid json or base64 string", http.StatusInternalServerError)
		return
	}

	cccData, err := base64.StdEncoding.DecodeString(r.CustomCheckConfiguration)
	if err != nil {
		log.Errorln("Webserver: Could not decode custom check configuration string for configuration push: ", err)
		http.Error(response, "invalid json or base64 string", http.StatusInternalServerError)
		return
	}

	if len(cfgData) == 0 {
		log.Errorln("Webserver: received empty configuration for configuration push: ", err)
		http.Error(response, "invalid json or base64 string", http.StatusInternalServerError)
		return
	}

	if err := w.Configuration.SaveConfiguration(cfgData); err != nil {
		log.Errorln("Webserver: ", err)
	}

	if err := w.Configuration.SaveConfiguration(cccData); err != nil {
		log.Errorln("Webserver: ", err)
	}

	if w.Reloader != nil {
		go w.Reloader.Reload()
	}
}

func (w *handler) handlerCsr(response http.ResponseWriter, request *http.Request) {
	log.Infoln("Webserver: openITCOCKPIT requests the CSR")

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
	log.Infoln("Webserver: CSR generated")
	if err := ioutil.WriteFile(w.Configuration.AutoSslCsrFile, csr, 0600); err != nil {
		log.Infoln("Webserver: could not store csr: ", err)
	}
	js, err := json.Marshal(csrResponse{Csr: string(csr)})
	if err != nil {
		log.Errorln("Webserver: Could not create json for csr: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}
	log.Debugln("Webserver: Send CSR: ", string(js))
	response.Header().Add("Content-Type", "application/json")
	if _, err := response.Write(js); err != nil {
		log.Errorln("Webserver: ", err)
	} else {
		log.Infoln("Webserver: CSR sent successfully")
	}
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

	log.Debugln("Webserver: received new certificate: ", string(body))

	crtReq := &updateCrtRequest{}
	if err := json.Unmarshal([]byte(body), crtReq); err != nil {
		log.Errorln("Webserver: Could not parse certificate update request: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := ioutil.WriteFile(w.Configuration.AutoSslCrtFile, []byte(crtReq.Signed), 0600); err != nil {
		log.Errorln("Webserver: Could not write certificate file: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := ioutil.WriteFile(w.Configuration.AutoSslCaFile, []byte(crtReq.CA), 0600); err != nil {
		log.Errorln("Webserver: Could not write ca certificate file: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Debugln("Webserver: Certificate update successful, start reload")

	if w.Reloader != nil {
		go w.Reloader.Reload()
	}
}

// Handler can be used by http.Server to handle http connections
func (w *handler) Handler() *mux.Router {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.router == nil {
		routes := mux.NewRouter()
		if log.GetLevel() == log.DebugLevel {
			log.Debugln("Webserver: Activate Handler Debug Middleware")
			routes.Use(debugMiddleware)
		}
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

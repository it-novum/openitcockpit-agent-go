package webserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
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
	StateInput      <-chan []byte
	PrometheusInput <-chan map[string]string
	Reloader        Reloader
	Configuration   *config.Configuration

	mtx             sync.RWMutex
	prometheusMtx   sync.RWMutex
	shutdown        chan struct{}
	state           []byte
	prometheusState map[string]string
	wg              sync.WaitGroup

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

func (w *handler) getPrometheusState() map[string]string {
	w.prometheusMtx.RLock()
	defer w.prometheusMtx.RUnlock()

	return w.prometheusState
}

func (w *handler) setPrometheusState(newState map[string]string) {
	w.prometheusMtx.Lock()
	defer w.prometheusMtx.Unlock()
	log.Debugln("Webserver Prometheus: set new exporter state")

	// Create a new map to have a copy
	// https://stackoverflow.com/a/23058707/11885414
	state := make(map[string]string, len(newState))
	for k, v := range newState {
		state[k] = v
	}

	w.prometheusState = state
}

func (w *handler) handleStatus(response http.ResponseWriter, _ *http.Request) {
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_, err := response.Write(w.getState())
	if err != nil {
		log.Errorln("Webserver: ", err)
	}
}

func (w *handler) handlePrometheusExporterStatus(response http.ResponseWriter, r *http.Request) {
	exporter := r.URL.Query().Get("exporter") // ?exporter=node_exporter

	if exporter == "" {
		// Return a list of all available exporters
		exporterNames := make([]string, 0)
		for _, e := range w.Configuration.PrometheusExporterConfiguration {
			exporterNames = append(exporterNames, e.Name)
		}
		response.Header().Add("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)

		exporterNamesJson, err := json.Marshal(exporterNames)
		if err != nil {
			_, err = response.Write([]byte("[]"))
		} else {
			_, err = response.Write(exporterNamesJson)
		}

		if err != nil {
			log.Errorln("Webserver Prometheus: ", err)
		}

		return
	}

	// Return the output of a specific exporter
	response.Header().Add("Content-Type", "text/plain")

	exporterState := w.getPrometheusState()
	if exporterState != nil {
		response.WriteHeader(http.StatusOK)

		if val, ok := exporterState[exporter]; ok {
			_, err := response.Write([]byte(val))
			if err != nil {
				log.Errorln("Webserver: ", err)
			}

			return
		}
	}

	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Unknown exporter"))
}

type configurationPush struct {
	Configuration                   string `json:"configuration"`
	CustomCheckConfiguration        string `json:"customcheck_configuration"`
	PrometheusExporterConfiguration string `json:"prometheus_exporter"`
}

func (w *handler) handleConfigRead(response http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()

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

	data = w.Configuration.ReadPrometheusExporterConfiguration()
	r.PrometheusExporterConfiguration = base64.StdEncoding.EncodeToString(data)

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
	defer func() {
		_ = request.Body.Close()
	}()

	if !w.Configuration.ConfigUpdate {
		http.Error(response, "config update is disabled", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(request.Body)
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

	prometheusData, err := base64.StdEncoding.DecodeString(r.PrometheusExporterConfiguration)
	if err != nil {
		log.Errorln("Webserver: Could not decode Prometheus Exporter configuration string for configuration push: ", err)
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

	if err := w.Configuration.SaveCustomCheckConfiguration(cccData); err != nil {
		log.Errorln("Webserver: ", err)
	}

	if err := w.Configuration.SavePrometheusExporterConfiguration(prometheusData); err != nil {
		log.Errorln("Webserver: ", err)
	}

	if w.Reloader != nil {
		// Reload Agent Instance via Reloader interface
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
	if err := os.WriteFile(w.Configuration.AutoSslCsrFile, csr, 0600); err != nil {
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
	defer func() {
		_ = request.Body.Close()
	}()

	log.Debugln("Webserver: Certificate update")

	body, err := io.ReadAll(request.Body)
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

	if err := os.WriteFile(w.Configuration.AutoSslCrtFile, []byte(crtReq.Signed), 0600); err != nil {
		log.Errorln("Webserver: Could not write certificate file: ", err)
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(w.Configuration.AutoSslCaFile, []byte(crtReq.CA), 0600); err != nil {
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
	w.prometheusMtx.Lock()
	defer w.prometheusMtx.Unlock()
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
		routes.Path("/prometheus").Methods("GET").HandlerFunc(w.handlePrometheusExporterStatus)
		routes.Path("/config").Methods("GET").HandlerFunc(w.handleConfigRead)
		routes.Path("/config").Methods("POST").HandlerFunc(w.handleConfigPush)
		routes.Path("/autotls").Methods("GET").HandlerFunc(w.handlerCsr)
		routes.Path("/autotls").Methods("POST").HandlerFunc(w.handlerUpdateCert)

		if w.Configuration.EnablePPROF {
			routes.Path("/debug/pprof/").HandlerFunc(pprof.Index)
			routes.Path("/debug/pprof/cmdline").HandlerFunc(pprof.Cmdline)
			routes.Path("/debug/pprof/profile").HandlerFunc(pprof.Profile)
			routes.Path("/debug/pprof/symbol").HandlerFunc(pprof.Symbol)
			routes.Path("/debug/pprof/trace").HandlerFunc(pprof.Trace)
			routes.Path("/debug/pprof/block").Handler(pprof.Handler("block"))
			routes.Path("/debug/pprof/goroutine").Handler(pprof.Handler("goroutine"))
			routes.Path("/debug/pprof/heap").Handler(pprof.Handler("heap"))
			routes.Path("/debug/pprof/threadcreate").Handler(pprof.Handler("threadcreate"))
		}

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
			case s := <-w.PrometheusInput:
				w.setPrometheusState(s)
			}
		}
	}()
}

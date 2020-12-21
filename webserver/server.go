package webserver

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	log "github.com/sirupsen/logrus"
)

// ReloadConfig for the webserver
type ReloadConfig struct {
	WebServer *config.WebServer
	TLS       *config.TLS
	BasicAuth *config.BasicAuth
	// reloadDone will be set by the reload func
	reloadDone chan struct{}
}

// Server handling for http, should be created by New
type Server struct {
	reload   chan *ReloadConfig
	shutdown chan struct{}

	stateInput          <-chan []byte
	configPushRecipient chan<- string

	server  *http.Server
	handler *handler

	wg sync.WaitGroup
}

func (s *Server) doReload(ctx context.Context, cfg *ReloadConfig) {
	log.Infoln("Webserver: Reload")
	newHandler := &handler{
		StateInput:          s.stateInput,
		ConfigPushRecipient: s.configPushRecipient,
		TLS:                 cfg.TLS,
		BasicAuthConfig:     cfg.BasicAuth,
	}
	newHandler.prepare()
	go newHandler.Run(ctx) // will be stopped by close()
	serverAddr := fmt.Sprintf("%s:%d", cfg.WebServer.Address, cfg.WebServer.Port)
	log.Debugln("Webserver: Listening to ", serverAddr)
	newServer := &http.Server{
		Addr:           serverAddr,
		Handler:        newHandler.Handler(),
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 30,
		IdleTimeout:    time.Second * 30,
		MaxHeaderBytes: 256 * 1024,
	}

	if cfg.TLS.AutoSslEnabled || (cfg.TLS.KeyFile != "" && cfg.TLS.CertificateFile != "") {
		log.Debugln("Webserver: TLS enabled")
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		certFilePath := cfg.TLS.CertificateFile
		keyFilePath := cfg.TLS.KeyFile
		caFilePath := ""
		if cfg.TLS.AutoSslEnabled {
			log.Debugln("Webserver: Using AutoSSL certificates")

			certFilePath = cfg.TLS.AutoSslCrtFile
			keyFilePath = cfg.TLS.AutoSslKeyFile
			caFilePath = cfg.TLS.AutoSslCaFile

			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		pem := bytes.Buffer{}

		certPem, err := ioutil.ReadFile(certFilePath)
		if err != nil {
			log.Fatalln("Webserver: Could not read server certificate: ", err)
		}
		pem.Write(certPem)
		pem.WriteByte('\n')
		keyPem, err := ioutil.ReadFile(keyFilePath)
		if err != nil {
			log.Fatalln("Webserver: Could not read server key: ", err)
		}

		if caFilePath != "" {
			pool, caPem, err := utils.CertPoolFromFiles(caFilePath)
			if err != nil {
				log.Fatalln("Webserver: ", err)
			}
			tlsConfig.ClientCAs = pool
			log.Debugln("Webserver: Loaded ca certificate")
			pem.Write(caPem)
		}

		cert, err := tls.X509KeyPair(pem.Bytes(), keyPem)
		if err != nil {
			log.Fatal("Webserver: Could not load tls certificate: ", err)
		}
		log.Debugln("Webserver: Loaded server cerificate")

		tlsConfig.Certificates = []tls.Certificate{cert}

		newServer.TLSConfig = tlsConfig
		newServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	s.close()
	s.handler = newHandler

	s.wg.Add(1)
	go func() {
		log.Infoln("Webserver: Starting http server")
		err := listenServe(newServer)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Webserver: ", err)
		}
		log.Debugln("Webserver: http listener stopped")
		s.wg.Done()
	}()

	s.server = newServer
	log.Debugln("Webserver: Reload complete")
	cfg.reloadDone <- struct{}{}
}

func (s *Server) close() {
	if s.server != nil {
		log.Debugln("Webserver: Stopping http server")
		s.server.Close()
		s.wg.Wait()
		s.server = nil
		log.Infoln("Webserver: Server stopped")
	}
	if s.handler != nil {
		log.Debugln("Webserver: Stopping handler")
		s.handler.Shutdown()
		s.handler = nil
		log.Debugln("Webserver: Handler stopped")
	}
}

func listenServe(server *http.Server) error {
	if server.TLSConfig != nil {
		return server.ListenAndServeTLS("", "")
	}
	return server.ListenAndServe()
}

// Reload webserver configuration
func (s *Server) Reload(reloadConfig *ReloadConfig) {
	done := make(chan struct{})
	reloadConfig.reloadDone = done
	s.reload <- reloadConfig
	<-done
}

// Shutdown webserver
func (s *Server) Shutdown() {
	close(s.shutdown)
	s.wg.Wait()
}

// New http server handler
func New(stateInput <-chan []byte, configPushRecipient chan<- string) *Server {
	return &Server{
		reload:              make(chan *ReloadConfig),
		shutdown:            make(chan struct{}),
		stateInput:          stateInput,
		configPushRecipient: configPushRecipient,
	}
}

// Run the server routine (should be run as go routine)
// You have to call Reload at least once to really start the webserver
func (s *Server) Run(ctx context.Context) {
	defer s.close()
	for {
		select {
		case <-ctx.Done():
			return
		case _, more := <-s.shutdown:
			if !more {
				return
			}
		case newConfig := <-s.reload:
			s.doReload(ctx, newConfig)
		}
	}
}

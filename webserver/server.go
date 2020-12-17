package webserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

type reloadConfig struct {
	web        *config.WebServer
	tls        *config.TLS
	reloadDone chan struct{}
}

// Server handling for http
type Server struct {
	server   *http.Server
	reload   chan *reloadConfig
	shutdown chan struct{}
	handler  *handler
	wg       sync.WaitGroup
}

func (s *Server) doReload(cfg *reloadConfig) {
	newServer := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.web.Address, cfg.web.Port),
		Handler:        s.handler.Handler(),
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 30,
		IdleTimeout:    time.Second * 30,
		MaxHeaderBytes: 256 * 1024,
	}

	if cfg.tls.AutoSslEnabled || (cfg.tls.KeyFile != "" && cfg.tls.CertificateFile != "") {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		certFilePath := cfg.tls.CertificateFile
		keyFilePath := cfg.tls.KeyFile
		caFilePath := ""
		if cfg.tls.AutoSslEnabled {
			certFilePath = cfg.tls.AutoSslCrtFile
			keyFilePath = cfg.tls.AutoSslKeyFile
			caFilePath = cfg.tls.AutoSslCaFile
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
		if err != nil {
			log.Fatal("Could not load tls certificate: ", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		if caFilePath != "" {
			pool := x509.NewCertPool()
			caBytes, err := ioutil.ReadFile(caFilePath)
			if err != nil {
				log.Fatal("Could not load tls ca certificate: ", err)
			}
			if !pool.AppendCertsFromPEM(caBytes) {
				log.Fatal("Could not parse ca certificate, probably not a valid PEM file")
			}
			tlsConfig.ClientCAs = pool
			tlsConfig.RootCAs = pool
		}
		tlsConfig.BuildNameToCertificate()
		newServer.TLSConfig = tlsConfig
		newServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	s.close()

	s.wg.Add(1)
	go func() {
		err := listenServe(newServer)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Webserver error: ", err)
		}
		s.wg.Done()
	}()

	s.server = newServer
	cfg.reloadDone <- struct{}{}
}

func (s *Server) close() {
	if s.server != nil {
		s.server.Close()
		s.wg.Wait()
		s.server = nil
	}
}

func listenServe(server *http.Server) error {
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	if server.TLSConfig != nil {
		listener = tls.NewListener(listener, server.TLSConfig)
	}
	defer listener.Close()
	return server.Serve(listener)
}

// Reload webserver configuration
func (s *Server) Reload(webConfig *config.WebServer, tlsConfig *config.TLS) {
	done := make(chan struct{})
	s.reload <- &reloadConfig{
		web:        webConfig,
		tls:        tlsConfig,
		reloadDone: done,
	}
	<-done
}

// Shutdown webserver
func (s *Server) Shutdown() {
	s.shutdown <- struct{}{}
	s.wg.Wait()
}

// New http server handler
func New(stateInput <-chan []byte, configPushRecipient chan<- string) *Server {
	return &Server{
		reload:   make(chan *reloadConfig),
		shutdown: make(chan struct{}),
		handler: &handler{
			StateInput:          stateInput,
			ConfigPushRecipient: configPushRecipient,
		},
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
		case <-s.shutdown:
			return
		case newConfig := <-s.reload:
			s.doReload(newConfig)
		}
	}
}

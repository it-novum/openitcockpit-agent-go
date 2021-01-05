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

type reloadConfig struct {
	Configuration *config.Configuration
	// reloadDone will be set by the reload func
	reloadDone chan struct{}
}

// Server handling for http, should be created by New
type Server struct {
	StateInput <-chan []byte

	reload   chan *reloadConfig
	shutdown chan struct{}

	server  *http.Server
	handler *handler

	wg sync.WaitGroup
}

func (s *Server) doReload(ctx context.Context, cfg *reloadConfig) {
	log.Infoln("Webserver: Reload")
	newHandler := &handler{
		StateInput:    s.StateInput,
		Configuration: cfg.Configuration,
	}
	newHandler.prepare()
	go newHandler.Run(ctx) // will be stopped by close()
	serverAddr := fmt.Sprintf("%s:%d", cfg.Configuration.Address, cfg.Configuration.Port)
	log.Debugln("Webserver: Listening to ", serverAddr)
	newServer := &http.Server{
		Addr:           serverAddr,
		Handler:        newHandler.Handler(),
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 30,
		IdleTimeout:    time.Second * 30,
		MaxHeaderBytes: 256 * 1024,
	}

	if cfg.Configuration.AutoSslEnabled || (cfg.Configuration.KeyFile != "" && cfg.Configuration.CertificateFile != "") {
		log.Debugln("Webserver: TLS enabled")
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		certFilePath := cfg.Configuration.CertificateFile
		keyFilePath := cfg.Configuration.KeyFile
		caFilePath := ""
		if cfg.Configuration.AutoSslEnabled {
			log.Debugln("Webserver: Using AutoSSL certificates")

			certFilePath = cfg.Configuration.AutoSslCrtFile
			keyFilePath = cfg.Configuration.AutoSslKeyFile
			caFilePath = cfg.Configuration.AutoSslCaFile

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
		newServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
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
func (s *Server) Reload(cfg *config.Configuration) {
	done := make(chan struct{})
	s.reload <- &reloadConfig{
		Configuration: cfg,
		reloadDone:    done,
	}
	<-done
}

// Shutdown webserver
func (s *Server) Shutdown() {
	close(s.shutdown)
	s.wg.Wait()
}

// Run the server routine (should NOT be run in a go routine)
// You have to call Reload at least once to really start the webserver
func (s *Server) Start(ctx context.Context) {
	s.shutdown = make(chan struct{})
	s.reload = make(chan *reloadConfig)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

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
	}()
}

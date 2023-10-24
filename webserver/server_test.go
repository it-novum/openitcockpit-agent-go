package webserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestServer(t *testing.T) {
	stateInput := make(chan []byte)
	srv := &Server{
		StateInput: stateInput,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv.Start(ctx)
	port := dynamicPort()
	srv.Reload(&config.Configuration{
		Port: port,
	})
	if !connectionTest("localhost", int(port), nil) {
		t.Error("server did not start correctly")
	}
	done := make(chan struct{})
	go func() {
		srv.Shutdown()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Error("timeout for Shutdown reached")
	}
}

func TestServerCancel(t *testing.T) {
	stateInput := make(chan []byte)
	srv := &Server{
		StateInput: stateInput,
	}
	ctx, cancel := context.WithCancel(context.Background())
	srv.Start(ctx)
	port := dynamicPort()
	srv.Reload(&config.Configuration{
		Port: port,
	})
	if !connectionTest("localhost", int(port), nil) {
		t.Error("server did not start correctly")
	}
	cancel()

	done := make(chan struct{})
	go func() {
		srv.wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Error("timeout for cancel reached")
	}
	if connectionTest("localhost", int(port), nil) {
		t.Error("server is still running")
	}
}

func TestServerTLS(t *testing.T) {
	crt, err := copyTestCertificates(false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(crt.tmpDir)
	}()

	stateInput := make(chan []byte)

	srv := &Server{
		StateInput: stateInput,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv.Start(ctx)

	port := dynamicPort()
	srv.Reload(&config.Configuration{
		Port:            port,
		KeyFile:         crt.keyPath,
		CertificateFile: crt.certPath,
	})
	if !connectionTest("localhost", int(port), crt) {
		t.Error("server did not start correctly")
	}

	done := make(chan struct{})
	go func() {
		srv.Shutdown()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Error("timeout for Shutdown reached")
	}
}

func TestServerAutoTLS(t *testing.T) {
	crt, err := copyTestCertificates(true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(crt.tmpDir)
	}()

	stateInput := make(chan []byte)

	srv := &Server{
		StateInput: stateInput,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv.Start(ctx)

	port := dynamicPort()
	srv.Reload(&config.Configuration{
		Port:           port,
		AutoSslEnabled: true,
		AutoSslFolder:  crt.tmpDir,
		AutoSslCrtFile: crt.certPath,
		AutoSslKeyFile: crt.keyPath,
		AutoSslCaFile:  crt.caCertPath,
	})
	if !connectionTest("localhost", int(port), crt) {
		t.Error("server did not start correctly")
	}

	done := make(chan struct{})
	go func() {
		srv.Shutdown()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Error("timeout for Shutdown reached")
	}
}

func TestServerAutoTLSBasicAuthRealClient(t *testing.T) {
	crt, err := copyTestCertificates(true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(crt.tmpDir)
	}()

	stateInput := make(chan []byte)

	srv := &Server{
		StateInput: stateInput,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv.Start(ctx)

	port := dynamicPort()
	srv.Reload(&config.Configuration{
		Port:           port,
		AutoSslEnabled: true,
		AutoSslFolder:  crt.tmpDir,
		AutoSslCrtFile: crt.certPath,
		AutoSslKeyFile: crt.keyPath,
		AutoSslCaFile:  crt.caCertPath,
		BasicAuth:      testBasicAuth,
	})
	stateInput <- []byte(`{"test": "tata"}`)

	caPool, _, err := utils.CertPoolFromFiles(crt.caCertPath)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := tls.LoadX509KeyPair(crt.certClientPath, crt.keyClientPath)
	if err != nil {
		t.Fatal(err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      caPool,
			Certificates: []tls.Certificate{cert},
		},
	}
	c := http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d", port), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(testBasicAuthUser, testBasicAuthPassword)
	if res, err := c.Do(req); err != nil {
		t.Error(err)
	} else {
		defer func() {
			_ = res.Body.Close()
		}()
		if body, err := io.ReadAll(res.Body); err != nil {
			t.Error(body)
		} else {
			sbody := string(body)
			if sbody != `{"test": "tata"}` {
				t.Error("unexpected body: ", sbody)
			}
		}
	}

	done := make(chan struct{})
	go func() {
		srv.Shutdown()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Error("timeout for Shutdown reached")
	}
}

func dynamicPort() int64 {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = l.Close()
	}()
	return int64(l.Addr().(*net.TCPAddr).Port)
}

func connectionTest(host string, port int, crt *certs) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	if crt != nil {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		if crt.caCertPath != "" {
			certPool := x509.NewCertPool()
			caPEM, err := os.ReadFile(crt.caCertPath)
			if err != nil {
				log.Fatal("could not read ca file: ", err)
			}
			certPool.AppendCertsFromPEM(caPEM)
			tlsConfig.RootCAs = certPool
			clientCert, err := tls.LoadX509KeyPair(crt.certClientPath, crt.keyClientPath)
			if err != nil {
				log.Println(err)
				return false
			}
			tlsConfig.Certificates = []tls.Certificate{
				clientCert,
			}
			tlsConn := tls.Client(conn, tlsConfig)
			if err := tlsConn.Handshake(); err != nil {
				log.Println(err)
				return false
			}
		} else {
			tlsConfig.InsecureSkipVerify = true
		}
	}
	return true
}

type certs struct {
	tmpDir         string
	certPath       string
	keyPath        string
	certClientPath string
	keyClientPath  string
	caCertPath     string
	caKeyPath      string
}

func copyTestCertificates(autoTLS bool) (*certs, error) {
	crt := &certs{}
	tmpDir, err := os.MkdirTemp(os.TempDir(), "*-test")
	crt.tmpDir = tmpDir
	ok := false
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			_ = os.RemoveAll(tmpDir)
		}
	}()
	_, filename, _, _ := runtime.Caller(0)
	templateCertDir := filepath.Join(filepath.Dir(filename), "..", "testdata", "certificates")

	crt.certPath = filepath.Join(tmpDir, "server.crt")
	crt.keyPath = filepath.Join(tmpDir, "server.key")
	copyMap := map[string]string{}
	if autoTLS {
		crt.caCertPath = filepath.Join(tmpDir, "ca.crt")
		crt.caKeyPath = filepath.Join(tmpDir, "ca.key")
		crt.certClientPath = filepath.Join(tmpDir, "client.crt")
		crt.keyClientPath = filepath.Join(tmpDir, "client.key")
		copyMap["server.crt"] = "server.crt"
		copyMap["server.key"] = "server.key"
		copyMap["ca.crt"] = "ca.crt"
		copyMap["ca.key"] = "ca.key"
		copyMap["client.crt"] = "client.crt"
		copyMap["client.key"] = "client.key"
	} else {
		copyMap["server_self.crt"] = "server.crt"
		copyMap["server_self.key"] = "server.key"
	}
	for src, dest := range copyMap {
		if err := utils.CopyFile(filepath.Join(templateCertDir, src), filepath.Join(tmpDir, dest)); err != nil {
			log.Fatalln("Copy failed: ", filepath.Join(templateCertDir, src), "->", filepath.Join(tmpDir, dest), ": ", err)
		}
	}
	ok = true
	return crt, nil
}

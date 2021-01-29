package webserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	crt, err := generateTestCertificates(false)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(crt.tmpDir)

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
	crt, err := generateTestCertificates(true)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(crt.tmpDir)

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
	crt, err := generateTestCertificates(true)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(crt.tmpDir)

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
		defer res.Body.Close()
		if body, err := ioutil.ReadAll(res.Body); err != nil {
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
	defer l.Close()
	return int64(l.Addr().(*net.TCPAddr).Port)
}

func connectionTest(host string, port int, crt *certs) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	if crt != nil {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		if crt.caCertPath != "" {
			certPool := x509.NewCertPool()
			caPEM, err := ioutil.ReadFile(crt.caCertPath)
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

func encodeAndSaveKey(dest string, key *rsa.PrivateKey) error {
	buf := &bytes.Buffer{}
	if err := pem.Encode(buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return fmt.Errorf("encode private key failed: %s", err)
	}
	if err := ioutil.WriteFile(dest, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("write key to file failed: %s", err)
	}
	return nil
}

func encodeAndSaveCert(dest string, derCert []byte) error {
	buf := &bytes.Buffer{}
	if err := pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCert,
	}); err != nil {
		return fmt.Errorf("encode certificate failed: %s", err)
	}
	if err := ioutil.WriteFile(dest, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("write certificate to file failed: %s", err)
	}
	return nil
}

func generateSelfSignedCertificate(destCertFilePath, destKeyFilePath string) error {
	bits := 4096
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("rsa key generate failed: %s", err)
	}

	tpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	derCert, err := x509.CreateCertificate(rand.Reader, &tpl, &tpl, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("create x509 certificate failed: %s", err)
	}

	if err := encodeAndSaveKey(destKeyFilePath, privateKey); err != nil {
		return err
	}

	return encodeAndSaveCert(destCertFilePath, derCert)
}

func generateCASignedCertificate(destCertFilePath, destKeyFilePath, destClientCertFilePath, destClientKeyFilePath, destCACertFilePath, destCAKeyFilePath string) error {
	bits := 4096
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("rsa key generate failed: %s", err)
	}
	privateClientKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("rsa key generate failed: %s", err)
	}
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("rsa key generate failed: %s", err)
	}

	ca := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "CA Name"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
	}
	caDer, err := x509.CreateCertificate(rand.Reader, &ca, &ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return fmt.Errorf("create x509 certificate failed: %s", err)
	}

	cert := x509.Certificate{
		SerialNumber:          big.NewInt(456),
		Subject:               pkix.Name{CommonName: "localhost"},
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	derCert, err := x509.CreateCertificate(rand.Reader, &cert, &ca, &privateKey.PublicKey, caPrivateKey)
	if err != nil {
		return fmt.Errorf("create x509 certificate failed: %s", err)
	}

	certClient := x509.Certificate{
		SerialNumber:          big.NewInt(457),
		Subject:               pkix.Name{CommonName: "test client"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	derCertClient, err := x509.CreateCertificate(rand.Reader, &certClient, &ca, &privateClientKey.PublicKey, caPrivateKey)
	if err != nil {
		return fmt.Errorf("create x509 certificate failed: %s", err)
	}

	if err := encodeAndSaveCert(destCACertFilePath, caDer); err != nil {
		return err
	}

	if err := encodeAndSaveKey(destCAKeyFilePath, caPrivateKey); err != nil {
		return err
	}

	if err := encodeAndSaveCert(destCertFilePath, derCert); err != nil {
		return err
	}

	if err := encodeAndSaveKey(destKeyFilePath, privateKey); err != nil {
		return err
	}

	if err := encodeAndSaveCert(destClientCertFilePath, derCertClient); err != nil {
		return err
	}

	if err := encodeAndSaveKey(destClientKeyFilePath, privateClientKey); err != nil {
		return err
	}

	return nil
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

func generateTestCertificates(autoTLS bool) (*certs, error) {
	crt := &certs{}
	tmpDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	crt.tmpDir = tmpDir
	ok := false
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			os.RemoveAll(tmpDir)
		}
	}()
	crt.certPath = filepath.Join(tmpDir, "server.crt")
	crt.keyPath = filepath.Join(tmpDir, "server.key")
	if autoTLS {
		crt.caCertPath = filepath.Join(tmpDir, "ca.crt")
		crt.caKeyPath = filepath.Join(tmpDir, "ca.key")
		crt.certClientPath = filepath.Join(tmpDir, "client.crt")
		crt.keyClientPath = filepath.Join(tmpDir, "client.key")
		if err := generateCASignedCertificate(crt.certPath, crt.keyPath, crt.certClientPath, crt.keyClientPath, crt.caCertPath, crt.caKeyPath); err != nil {
			return nil, err
		}
	} else {
		if err := generateSelfSignedCertificate(crt.certPath, crt.keyPath); err != nil {
			return nil, err
		}
	}
	ok = true
	return crt, nil
}

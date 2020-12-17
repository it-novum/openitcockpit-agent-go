package webserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

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

func connectionTest(host string, port int) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

func generateSelfSignedCertificate(destCertFilePath, destKeyFilePath string) error {
	bits := 4096
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("rsa key generate failed: %s", err)
	}

	tpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "169.264.169.254"},
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

	buf := &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCert,
	})
	if err != nil {
		return fmt.Errorf("encode certificate failed: %s", err)
	}

	pemCert := buf.Bytes()

	buf = &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return fmt.Errorf("encode private key failed: %s", err)
	}
	pemKey := buf.Bytes()

	if err := ioutil.WriteFile(destCertFilePath, pemCert, 0666); err != nil {
		return fmt.Errorf("write certificate to file failed: %s", err)
	}
	if err := ioutil.WriteFile(destKeyFilePath, pemKey, 0666); err != nil {
		return fmt.Errorf("write key to file failed: %s", err)
	}

	return nil
}

func TestServer(t *testing.T) {
	stateInput := make(chan []byte)
	configPush := make(chan string)
	srv := New(stateInput, configPush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		srv.Run(ctx)
		done <- struct{}{}
	}()
	port := dynamicPort()
	srv.Reload(&config.WebServer{
		Address: "",
		Port:    port,
	}, &config.TLS{})
	if !connectionTest("localhost", int(port)) {
		t.Error("server did not start correctly")
	}
	srv.Shutdown()
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Error("timeout for Shutdown reached")
	}
}

func TestServerCancel(t *testing.T) {
	stateInput := make(chan []byte)
	configPush := make(chan string)
	srv := New(stateInput, configPush)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		srv.Run(ctx)
		done <- struct{}{}
	}()
	port := dynamicPort()
	srv.Reload(&config.WebServer{
		Address: "",
		Port:    port,
	}, &config.TLS{})
	if !connectionTest("localhost", int(port)) {
		t.Error("server did not start correctly")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Error("timeout for cancel reached")
	}
	if connectionTest("localhost", int(port)) {
		t.Error("server is still running")
	}
}

func TestServerTLS(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	certPath := path.Join(tmpDir, "server.crt")
	keyPath := path.Join(tmpDir, "server.key")
	if err := generateSelfSignedCertificate(certPath, keyPath); err != nil {
		t.Fatal(err)
	}

	stateInput := make(chan []byte)
	configPush := make(chan string)

	srv := New(stateInput, configPush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		srv.Run(ctx)
		done <- struct{}{}
	}()

	port := dynamicPort()
	srv.Reload(&config.WebServer{
		Address: "",
		Port:    port,
	}, &config.TLS{
		KeyFile:         keyPath,
		CertificateFile: certPath,
	})
	if !connectionTest("localhost", int(port)) {
		t.Error("server did not start correctly")
	}

	srv.Shutdown()
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Error("timeout for Shutdown reached")
	}
}

// TODO test autossl

package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// CertPoolFromFiles reads listed file names and returns the certpool for ca or other usage
func CertPoolFromFiles(files ...string) (*x509.CertPool, []byte, error) {
	p := x509.NewCertPool()
	pem := bytes.Buffer{}

	for _, fileName := range files {
		bytes, err := os.ReadFile(fileName)
		if err != nil {
			return nil, nil, err
		}
		if !p.AppendCertsFromPEM(bytes) {
			return nil, nil, fmt.Errorf("Not a valid pem encoded certificate %s", fileName)
		}
		pem.Write(bytes)
		pem.WriteByte('\n')
	}

	return p, pem.Bytes(), nil
}

// GeneratePrivateKeyIfNotExists checks for keyFile and if it does not exist generates a rsa 4096 bits key
func GeneratePrivateKeyIfNotExists(keyFile string) error {
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		key, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return err
		}

		pemBytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return err
		}

		pemData := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pemBytes,
		})
		if err := os.WriteFile(keyFile, pemData, 0600); err != nil {
			return err
		}

		// Make sure that the private key can only be readed by the current user
		// We have to use utils.Chmod because on Windows Systems golang was to lazy to implement propper Windows filesystem permissions
		if err := Chmod(keyFile, 0600); err != nil {
			log.Errorln("Could not set file permissions to private key file to current user only")
			return err
		}
	}
	return nil
}

// CSRFromKeyFile reads keyFile and generates a csr in PEM format
func CSRFromKeyFile(keyFile, subject string) ([]byte, error) {
	pemData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	pemBlock, _ := pem.Decode(pemData)
	if pemBlock == nil {
		return nil, fmt.Errorf("key file does not contain any valid pem block")
	}
	key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	if subject == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		subject = hostname
	}

	req := &x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,
		Subject: pkix.Name{
			CommonName: "subject",
		},
		DNSNames: []string{subject},
	}

	der, err := x509.CreateCertificateRequest(rand.Reader, req, key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: der,
	}), nil
}

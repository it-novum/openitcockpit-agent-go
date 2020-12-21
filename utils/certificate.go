package utils

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// CertPoolFromFiles reads listed file names and returns the certpool for ca or other usage
func CertPoolFromFiles(files ...string) (*x509.CertPool, []byte, error) {
	p := x509.NewCertPool()
	pem := bytes.Buffer{}

	for _, fileName := range files {
		bytes, err := ioutil.ReadFile(fileName)
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

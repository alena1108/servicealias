package kubernetesclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

const (
	caLocation = "/etc/kubernetes/ssl/ca.pem"
)

var (
	token  string
	caData []byte
)

func Init() error {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("Failed to read token from stdin: %v", err)
	}
	token = strings.TrimSpace(string(bytes))
	if token == "" {
		return errors.New("No token passed in from stdin")
	}

	log.Infof("token is %s", token)
	caData, err = ioutil.ReadFile(caLocation)
	if err != nil {
		return fmt.Errorf("Failed to read CA cert %s: %v", caLocation, err)
	}
	log.Infof("ca data is %s", caData)

	return nil
}

func GetAuthorizationHeader() string {
	log.Infof("Bearer token is %s", token)
	return fmt.Sprintf("Bearer %s", token)
}

func GetTLSClientConfig() *tls.Config {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)
	return &tls.Config{
		RootCAs: certPool,
	}
}

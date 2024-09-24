package executor

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func (e *Executor) HTTPSCheck(url, caPath string, insecure bool, retryLimit, interval int) error {

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	cafile, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("error while checking url %s => %v", url, err)
	}

	if ok := rootCAs.AppendCertsFromPEM(cafile); !ok {
		logrus.Infoln("No certs appended, using system certs only to call", url)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
		ClientCAs:          rootCAs,
	}

	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request to check url '%s' => '%v'", url, err)
	}

	for retry := 0; retry < retryLimit; retry++ {
		logrus.Infof("trying to query %s", url)
		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return fmt.Errorf("timeout check against %s took too long", url)
}

func (e *Executor) HTTPCheck(url string) error {

	return nil
}

func (e *Executor) KubectlCheck() error {
	return nil
}

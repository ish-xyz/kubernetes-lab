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

func (e *Executor) URLCheck(url, caPath string, insecure, isTLS bool, maxRetries, interval int) error {

	var client *http.Client

	if isTLS {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Only read and append the custom CA if caPath is provided
		if caPath != "" {
			caFile, err := os.ReadFile(caPath)
			if err != nil {
				return fmt.Errorf("error reading CA file %s: %v", caPath, err)
			}

			if ok := rootCAs.AppendCertsFromPEM(caFile); !ok {
				return fmt.Errorf("failed to append CA certificate from %s", caPath)
			}
			logrus.Infof("Custom CA certificate appended from %s", caPath)
		} else {
			logrus.Infoln("No custom CA provided, using system certs only")
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: insecure,
			RootCAs:            rootCAs, // Use RootCAs instead of ClientCAs
		}

		tr := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request to check url '%s': %v", url, err)
	}

	for retry := 0; retry < maxRetries; retry++ {
		logrus.Infof("Attempting to query %s (attempt %d/%d)", url, retry+1, maxRetries)
		resp, err := client.Do(req)
		if err != nil {
			logrus.Errorf("Error querying %s: %v", url, err)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		defer resp.Body.Close()

		logrus.Infof("Response status code: %d", resp.StatusCode)
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}

	return fmt.Errorf("timeout: check against %s failed after %d attempts", url, maxRetries)
}

func (e *Executor) KubectlCheck() error {
	return nil
}

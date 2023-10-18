package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	harness "github.com/harness/ff-golang-server-sdk/client"
	"github.com/harness/ff-golang-server-sdk/evaluation"
)

var (
	flagName string = getEnvOrDefault("FF_FLAG_NAME", "harnessappdemodarkmode")
	sdkKey   string = getEnvOrDefault("FF_API_KEY", "change me")
)

func main() {
	log.Println("Harness SDK Getting Started")

	certPool, err := loadCertificates([]string{"path to PEM", "path to PEM"})
	if err != nil {
		log.Printf("Failed to parse PEM files: `%s`\n", err)
	}

	// Create a custom TLS configuration and use the CA pool.
	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := http.Client{Transport: transport}

	// Create a feature flag client and wait for it to successfully initialize
	startTime := time.Now()

	// Note that this code uses ffserver hostname as an example, likely you'll have your own hostname or IP.
	// You should ensure the endpoint is returning a cert with valid SANs configured for the host/IP.
	client, err := harness.NewCfClient(sdkKey, harness.WithEventsURL("https://ffserver:8001/api/1.0"), harness.WithURL("https://ffserver:8001/api/1.0"), harness.WithWaitForInitialized(true), harness.WithHTTPClient(&httpClient))

	elapsedTime := time.Since(startTime)
	log.Printf("Took '%v' seconds to get a client initialization result ", elapsedTime.Seconds())

	if err != nil {
		log.Printf("Client failed to initialize: `%s`\n", err)
	}

	defer func() {
		err := client.Close()
		if err != nil {
			return
		}
	}()

	// Create a target (different targets can get different results based on rules)
	target := evaluation.Target{
		Identifier: "HT_1",
		Name:       "Harness_Target_1",
		Attributes: &map[string]interface{}{"email": "demo@harness.io"},
	}

	// Loop forever reporting the state of the flag
	for {
		resultBool, err := client.BoolVariation(flagName, &target, false)
		if err != nil {
			log.Printf("failed to get evaluation: %v ", err)
		}
		log.Printf("Flag variation %v\n", resultBool)

		time.Sleep(10 * time.Second)
	}

}

func getEnvOrDefault(key, defaultStr string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultStr
	}
	return value
}

// Load certificates from PEM files
func loadCertificates(filePaths []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, ca := range filePaths {
		var caBytes []byte
		var err error
		caBytes, err = os.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate from file: %w", err)
		}

		if !pool.AppendCertsFromPEM(caBytes) {
			return nil, fmt.Errorf("could not append CA certificate")
		}
	}
	return pool, nil
}

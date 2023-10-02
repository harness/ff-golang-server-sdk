package main

import (
	"log"
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

	// Create a feature flag client and wait for it to successfully initialize
	startTime := time.Now()
	client, err := harness.NewCfClient(sdkKey, harness.WithWaitForInitialized(false))
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
			log.Fatal("failed to get evaluation: ", err)
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

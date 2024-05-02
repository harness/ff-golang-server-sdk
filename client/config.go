package client

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/harness/ff-golang-server-sdk/cache"
	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/harness/ff-golang-server-sdk/storage"
	"github.com/harness/ff-golang-server-sdk/stream"
	"github.com/harness/ff-golang-server-sdk/types"
	"github.com/hashicorp/go-retryablehttp"
)

type config struct {
	url                    string
	eventsURL              string
	pullInterval           uint // in seconds
	Cache                  cache.Cache
	Store                  storage.Storage
	Logger                 logger.Logger
	httpClient             *http.Client
	authHttpClient         *http.Client
	enableStream           bool
	enableStore            bool
	target                 evaluation.Target
	eventStreamListener    stream.EventStreamListener
	enableAnalytics        bool
	proxyMode              bool
	waitForInitialized     bool
	maxAuthRetries         int
	authRetryStrategy      *backoff.ExponentialBackOff
	streamingRetryStrategy *backoff.ExponentialBackOff
	sleeper                types.Sleeper
	apiConfig              *apiConfiguration
}

type apiConfiguration struct {
	segmentRulesV2QueryParam rest.SegmentRulesV2QueryParam
}

func (a *apiConfiguration) GetSegmentRulesV2QueryParam() *rest.SegmentRulesV2QueryParam {
	return &a.segmentRulesV2QueryParam
}

func newDefaultConfig(log logger.Logger) *config {
	defaultCache, _ := cache.NewLruCache(10000, log) // size of cache
	var defaultStore storage.Storage
	if _, present := os.LookupEnv("DISABLE_LOCAL_CACHE"); !present {
		defaultStore = storage.NewFileStore("defaultProject", storage.GetHarnessDir(log), log)
	}

	// Authentication uses a default http client + timeout as we have our own custom retry logic for authentication.
	const requestTimeout = time.Second * 30
	authHttpClient := &http.Client{}
	authHttpClient.Timeout = requestTimeout

	// Remaining requests use a go-retryablehttp client to handle retries.
	requestHttpClient := retryablehttp.NewClient()
	requestHttpClient.Logger = logger.NewRetryableLogger(log)
	requestHttpClient.RetryMax = 10

	// Assign a custom ErrorHandler. By default, the go-retryablehttp library doesn't return the final
	// network error from the server but instead reports that it has exhausted all retry attempts.
	requestHttpClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		message := ""
		if resp != nil {
			message = fmt.Sprintf("Error after '%d' connection attempts: '%s'", numTries, resp.Status)
		}

		// In practice, the error is usually nil and the response is used, but include this for any
		// edge cases.
		if err != nil {
			message = fmt.Sprintf("Error after %d connection attempts: %v\n", numTries, err)
		}

		return resp, fmt.Errorf(message)
	}

	apiConfig := &apiConfiguration{
		segmentRulesV2QueryParam: "v2",
	}

	return &config{
		url:             "https://config.ff.harness.io/api/1.0",
		eventsURL:       "https://events.ff.harness.io/api/1.0",
		pullInterval:    60,
		Cache:           defaultCache,
		Store:           defaultStore,
		Logger:          log,
		authHttpClient:  authHttpClient,
		httpClient:      requestHttpClient.StandardClient(),
		enableStream:    true,
		enableStore:     true,
		enableAnalytics: true,
		proxyMode:       false,
		// Indicate that we should retry forever by default
		maxAuthRetries:         -1,
		authRetryStrategy:      getDefaultExpBackoff(),
		streamingRetryStrategy: getDefaultExpBackoff(),
		sleeper:                &types.RealClock{},
		apiConfig:              apiConfig,
	}
}

func getDefaultExpBackoff() *backoff.ExponentialBackOff {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.InitialInterval = 1 * time.Second
	exponentialBackOff.MaxInterval = 1 * time.Minute
	exponentialBackOff.Multiplier = 2.0
	return exponentialBackOff
}

package client

import (
	"fmt"
	"github.com/harness/ff-golang-server-sdk/cache"
	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/storage"
	"github.com/harness/ff-golang-server-sdk/stream"
	"github.com/harness/ff-golang-server-sdk/types"
	"github.com/hashicorp/go-retryablehttp"
	"net/http"
	"os"
	"time"
)

type config struct {
	url                 string
	eventsURL           string
	pullInterval        uint // in seconds
	Cache               cache.Cache
	Store               storage.Storage
	Logger              logger.Logger
	httpClient          *http.Client
	authHttpClient      *http.Client
	enableStream        bool
	enableStore         bool
	target              evaluation.Target
	eventStreamListener stream.EventStreamListener
	enableAnalytics     bool
	proxyMode           bool
	waitForInitialized  bool
	maxAuthRetries      int
	sleeper             types.Sleeper
}

func newDefaultConfig(log logger.Logger) *config {
	defaultCache, _ := cache.NewLruCache(10000, log) // size of cache
	var defaultStore storage.Storage
	if _, present := os.LookupEnv("DISABLE_LOCAL_CACHE"); !present {
		defaultStore = storage.NewFileStore("defaultProject", storage.GetHarnessDir(log), log)
	}

	const requestTimeout = time.Second * 30

	// Authentication uses a default http client + timeout as we have our own custom retry logic for authentication.
	authHttpClient := http.DefaultClient
	authHttpClient.Timeout = requestTimeout

	// Remaining requests use a go-retryablehttp client to handle retries.
	requestHttpClient := retryablehttp.NewClient()
	requestHttpClient.Logger = logger.NewRetryableLogger(log)
	requestHttpClient.RetryMax = 1

	// Assign a custom ErrorHandler. By default, the go-retryablehttp library doesn't return the final
	// network error from the server but instead reports that it has exhausted all retry attempts.
	requestHttpClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		message := ""
		if resp != nil {
			message = fmt.Sprintf("Error after '%d' connection attempts: '%s'", numTries, resp.Status)
		}

		if err != nil {
			fmt.Printf("Error after %d connection attempts: %v\n", numTries, err)
		}

		customError := fmt.Errorf(message)

		return resp, customError
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
		maxAuthRetries: -1,
		sleeper:        &types.RealClock{},
	}
}

package client

import (
	"github.com/cenkalti/backoff/v4"
	"github.com/harness/ff-golang-server-sdk/cache"
	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/storage"
	"github.com/harness/ff-golang-server-sdk/stream"
	"github.com/harness/ff-golang-server-sdk/types"
	"net/http"
)

// ConfigOption is used as return value for advanced client configuration
// using options pattern
type ConfigOption func(config *config)

// WithAnalyticsEnabled en/disable cache and analytics data being sent.
func WithAnalyticsEnabled(val bool) ConfigOption {
	return func(config *config) {
		config.enableAnalytics = val
	}
}

// WithURL set baseUrl for communicating with ff server
func WithURL(url string) ConfigOption {
	return func(config *config) {
		config.url = url
	}
}

// WithEventsURL set eventsURL for communicating with ff server
func WithEventsURL(url string) ConfigOption {
	return func(config *config) {
		config.eventsURL = url
	}
}

// WithPullInterval set pulling interval in minutes
func WithPullInterval(interval uint) ConfigOption {
	return func(config *config) {
		config.pullInterval = interval
	}
}

// WithCache set custom cache or predefined one from cache package
func WithCache(cache cache.Cache) ConfigOption {
	return func(config *config) {
		config.Cache = cache
		// functional options order of execution can be changed by user
		// and we need to attach logger again
		config.Cache.SetLogger(config.Logger)
	}
}

// WithStore set custom storage or predefined one from storage package
func WithStore(store storage.Storage) ConfigOption {
	return func(config *config) {
		config.Store = store
		// functional options order of execution can be changed by user
		// and we need to attach logger again
		config.Store.SetLogger(config.Logger)
	}
}

// WithLogger set custom logger used in main application
func WithLogger(logger logger.Logger) ConfigOption {
	return func(config *config) {
		config.Logger = logger
	}
}

// WithStreamEnabled set stream on or off
func WithStreamEnabled(val bool) ConfigOption {
	return func(config *config) {
		config.enableStream = val
	}
}

// WithStoreEnabled set store on or off
func WithStoreEnabled(val bool) ConfigOption {
	return func(config *config) {
		config.enableStore = val
	}
}

// WithHTTPClient set auth and http client for use in interactions with ff server
func WithHTTPClient(client *http.Client) ConfigOption {
	return func(config *config) {
		config.authHttpClient = client
		config.httpClient = client
	}
}

// WithTarget sets target
func WithTarget(target evaluation.Target) ConfigOption {
	return func(config *config) {
		config.target = target
	}
}

// WithEventStreamListener configures the SDK to forward Events from the Feature
// Flag server to the passed EventStreamListener
func WithEventStreamListener(e stream.EventStreamListener) ConfigOption {
	return func(config *config) {
		config.eventStreamListener = e
	}
}

// WithProxyMode should be used when the SDK is being used inside the ff proxy to control the cache and handle sse events
func WithProxyMode(b bool) ConfigOption {
	return func(config *config) {
		config.proxyMode = b
	}
}

// WithWaitForInitialized configures the SDK to block the thread until initialization succeeds or fails
func WithWaitForInitialized(b bool) ConfigOption {
	return func(config *config) {
		config.waitForInitialized = b
	}
}

// WithMaxAuthRetries sets how many times the SDK will retry if authentication fails
func WithMaxAuthRetries(i int) ConfigOption {
	return func(config *config) {
		config.maxAuthRetries = i
	}
}

// WithRetryStrategy WiWithRetryStrategy sets the backoff and retry strategy for regular http clients used by the SDK, i.e. not the go-retryablehttp client
// Mainly used for testing purposes, as the SDKs default backoff strategy should be sufficient for most if not all scenarios.
func WithRetryStrategy(retryStrategy *backoff.ExponentialBackOff) ConfigOption {
	return func(config *config) {
		config.retryStrategy = retryStrategy
	}
}

// WithSleeper is used to aid in testing functionality that sleeps
func WithSleeper(sleeper types.Sleeper) ConfigOption {
	return func(config *config) {
		config.sleeper = sleeper
	}
}

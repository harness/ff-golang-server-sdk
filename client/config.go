package client

import (
	"net/http"

	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/stream"

	"github.com/harness/ff-golang-server-sdk/cache"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/storage"
	"github.com/hashicorp/go-retryablehttp"
)

type config struct {
	url                 string
	eventsURL           string
	pullInterval        uint // in seconds
	Cache               cache.Cache
	Store               storage.Storage
	Logger              logger.Logger
	httpClient          *http.Client
	enableStream        bool
	enableStore         bool
	target              evaluation.Target
	eventStreamListener stream.EventStreamListener
	enableAnalytics     bool
}

func newDefaultConfig(log logger.Logger) *config {
	defaultCache, _ := cache.NewLruCache(10000, log) // size of cache
	defaultStore := storage.NewFileStore("defaultProject", storage.GetHarnessDir(log), log)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	retryClient.Logger = logger.NewRetryableLogger(log)

	return &config{
		url:             "https://config.ff.harness.io/api/1.0",
		eventsURL:       "https://events.ff.harness.io/api/1.0",
		pullInterval:    60,
		Cache:           defaultCache,
		Store:           defaultStore,
		Logger:          log,
		httpClient:      retryClient.StandardClient(),
		enableStream:    true,
		enableStore:     true,
		enableAnalytics: true,
	}
}

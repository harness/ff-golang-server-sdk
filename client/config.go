package client

import (
	"log"
	"net/http"

	"github.com/drone/ff-golang-server-sdk.v0/evaluation"

	"github.com/drone/ff-golang-server-sdk/cache"
	"github.com/drone/ff-golang-server-sdk/logger"
	"github.com/drone/ff-golang-server-sdk/storage"
	"github.com/hashicorp/go-retryablehttp"
)

type config struct {
	url          string
	pullInterval uint // in minutes
	Cache        cache.Cache
	Store        storage.Storage
	Logger       logger.Logger
	httpClient   *http.Client
	enableStream bool
	enableStore  bool
	target       evaluation.Target
}

func newDefaultConfig() *config {
	defaultLogger, err := logger.NewZapLogger(false)
	if err != nil {
		log.Printf("Error creating zap logger instance, %v", err)
	}
	defaultCache, _ := cache.NewLruCache(10000, defaultLogger) // size of cache
	defaultStore := storage.NewFileStore("defaultProject", storage.GetHarnessDir(), defaultLogger)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10

	return &config{
		url:          "http://localhost:7999/api/1.0",
		pullInterval: 1,
		Cache:        defaultCache,
		Store:        defaultStore,
		Logger:       defaultLogger,
		httpClient:   retryClient.StandardClient(),
		enableStream: true,
		enableStore:  true,
	}
}

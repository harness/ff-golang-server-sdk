package client

import (
	"log"

	"github.com/drone/ff-golang-server-sdk.v0/cache"
	"github.com/drone/ff-golang-server-sdk.v0/logger"
	"github.com/drone/ff-golang-server-sdk.v0/storage"
)

type config struct {
	url          string
	pullInterval uint // in minutes
	Cache        cache.Cache
	Store        storage.Storage
	Logger       logger.Logger
	enableStream bool
}

func newDefaultConfig() *config {
	defaultLogger, err := logger.NewZapLogger(false)
	if err != nil {
		log.Printf("Error creating zap logger instance, %v", err)
	}
	defaultCache, _ := cache.NewLruCache(10000, defaultLogger) // size of cache
	defaultStore := storage.NewFileStore("defaultProject", storage.GetHarnessDir(), defaultLogger)

	return &config{
		url:          "http://localhost:7999/api/1.0",
		pullInterval: 1,
		Cache:        defaultCache,
		Store:        defaultStore,
		Logger:       defaultLogger,
		enableStream: true,
	}
}

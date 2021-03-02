package client

import (
	"github.com/wings-software/ff-client-sdk-go/pkg/cache"
	"github.com/wings-software/ff-client-sdk-go/pkg/logger"
	"github.com/wings-software/ff-client-sdk-go/pkg/storage"
	"log"
)

type Config struct {
	url          string
	pullInterval uint // in minutes
	Cache        cache.Cache
	Store        storage.Storage
	Logger       logger.Logger
	enableStream bool
}

func NewDefaultConfig() *Config {
	defaultLogger, err := logger.NewZapLogger(false)
	if err != nil {
		log.Printf("Error creating zap logger instance, %v", err)
	}
	defaultCache, _ := cache.NewLruCache(10000, defaultLogger) // size of cache
	defaultStore := storage.NewFileStore("defaultProject", storage.GetHarnessDir(), defaultLogger)

	return &Config{
		url:          "http://localhost:7999/api/1.0",
		pullInterval: 1,
		Cache:        defaultCache,
		Store:        defaultStore,
		Logger:       defaultLogger,
		enableStream: true,
	}
}

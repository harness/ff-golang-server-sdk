package client

import (
	"github.com/wings-software/ff-client-sdk-go/cache"
	"github.com/wings-software/ff-client-sdk-go/logger"
	"github.com/wings-software/ff-client-sdk-go/storage"
)

type ConfigOption func(config *Config)

func WithUrl(url string) ConfigOption {
	return func(config *Config) {
		config.url = url
	}
}

func WithPullInterval(interval uint) ConfigOption {
	return func(config *Config) {
		config.pullInterval = interval
	}
}

func WithCache(cache cache.Cache) ConfigOption {
	return func(config *Config) {
		config.Cache = cache
		// functional options order of execution can be changed by user
		// and we need to attach logger again
		config.Cache.SetLogger(config.Logger)
	}
}

func WithStore(store storage.Storage) ConfigOption {
	return func(config *Config) {
		config.Store = store
		// functional options order of execution can be changed by user
		// and we need to attach logger again
		config.Store.SetLogger(config.Logger)
	}
}

func WithLogger(logger logger.Logger) ConfigOption {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithStreamEnabled(val bool) ConfigOption {
	return func(config *Config) {
		config.enableStream = val
	}
}

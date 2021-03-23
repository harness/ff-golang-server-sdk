package client

import (
	"github.com/drone/ff-golang-server-sdk.v0/cache"
	"github.com/drone/ff-golang-server-sdk.v0/logger"
	"github.com/drone/ff-golang-server-sdk.v0/storage"
)

// ConfigOption is used as return value for advanced client configuration
// using options pattern
type ConfigOption func(config *config)

// WithURL set baseUrl for communicating with ff server
func WithURL(url string) ConfigOption {
	return func(config *config) {
		config.url = url
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

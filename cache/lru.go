package cache

import (
	"github.com/drone/ff-golang-server-sdk.v1/logger"
	lru "github.com/hashicorp/golang-lru"

	"reflect"
	"time"
)

// LRUCache is thread-safe LAST READ USED Cache
type LRUCache struct {
	*lru.Cache
	logger     logger.Logger
	lastUpdate time.Time
}

//NewLruCache creates a new LRU instance
func NewLruCache(size int, logger logger.Logger) (*LRUCache, error) {
	cache, err := lru.New(size)
	if err != nil {
		logger.Errorf("Error initializing LRU cache, err: %v", err)
		return nil, err
	}
	logger.Infof("Cache successfully initialized with size: %d", size)
	return &LRUCache{
		Cache:  cache,
		logger: logger,
	}, nil
}

func (lru *LRUCache) getTime() time.Time {
	return time.Now()
}

// Set a new value if it is different from the previous one.
// Returns true if an eviction occurred.
func (lru *LRUCache) Set(key interface{}, value interface{}) (evicted bool) {
	prev, _ := lru.Get(key)
	if !reflect.DeepEqual(prev, value) {
		add := lru.Cache.Add(key, value)
		lru.lastUpdate = lru.getTime()
		lru.logger.Debugf("cache value changed for key %s with value %v", key, value)
		return add
	}
	return false
}

// Contains checks if a key is in the cache
func (lru *LRUCache) Contains(key interface{}) bool {
	return lru.Cache.Contains(key)
}

// Get looks up a key's value from the cache.
func (lru *LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	return lru.Cache.Get(key)
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (lru *LRUCache) Keys() []interface{} {
	return lru.Cache.Keys()
}

// Len returns the number of items in the cache.
func (lru *LRUCache) Len() int {
	return lru.Cache.Len()
}

// Purge is used to completely clear the cache.
func (lru *LRUCache) Purge() {
	lru.Cache.Purge()
	lru.lastUpdate = lru.getTime()
}

// Remove removes the provided key from the cache.
func (lru *LRUCache) Remove(key interface{}) (present bool) {
	present = lru.Cache.Remove(key)
	lru.lastUpdate = lru.getTime()
	if present {
		lru.logger.Debugf("Cache item successfully removed %v", key)
	}
	return
}

// Resize changes the cache size.
func (lru *LRUCache) Resize(size int) (evicted int) {
	return lru.Cache.Resize(size)
}

// Updated lastUpdate information
func (lru *LRUCache) Updated() time.Time {
	return lru.lastUpdate
}

// SetLogger set logger
func (lru LRUCache) SetLogger(logger logger.Logger) {
	lru.logger = logger
}

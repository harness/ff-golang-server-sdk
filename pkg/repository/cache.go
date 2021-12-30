package repository

import (
	"github.com/harness/ff-golang-server-sdk/log"
	lru "github.com/hashicorp/golang-lru"
)

// Cache wrapper to integrate any 3rd party implementation
type Cache interface {
	Set(key interface{}, value interface{}) (evicted bool)
	Contains(key interface{}) bool
	Get(key interface{}) (value interface{}, ok bool)
	Keys() []interface{}
	Len() int
	Remove(key interface{}) (present bool)
}

// LRUCache is thread-safe LAST READ USED Cache
type LRUCache struct {
	*lru.Cache
}

var _ Cache = &LRUCache{}

//NewLruCache creates a new LRU instance
func NewLruCache(size int) (LRUCache, error) {
	cache, err := lru.New(size)
	if err != nil {
		log.Errorf("Error initializing LRU cache, err: %v", err)
		return LRUCache{}, err
	}
	log.Infof("Cache successfully initialized with size: %d", size)
	return LRUCache{
		cache,
	}, nil
}

// Set a new value if it is different from the previous one.
// Returns true if an eviction occurred.
func (lru LRUCache) Set(key interface{}, value interface{}) (evicted bool) {
	add := lru.Cache.Add(key, value)
	log.Debugf("cache value changed for key %s with value %v", key, value)
	return add
}

// Contains checks if a key is in the cache
func (lru LRUCache) Contains(key interface{}) bool {
	return lru.Cache.Contains(key)
}

// Get looks up a key's value from the cache.
func (lru LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	return lru.Cache.Get(key)
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (lru LRUCache) Keys() []interface{} {
	return lru.Cache.Keys()
}

// Len returns the number of items in the cache.
func (lru LRUCache) Len() int {
	return lru.Cache.Len()
}

// Remove removes the provided key from the cache.
func (lru LRUCache) Remove(key interface{}) (present bool) {
	present = lru.Cache.Remove(key)
	if present {
		log.Debugf("Cache item successfully removed %v", key)
	}
	return
}

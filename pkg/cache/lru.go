package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/wings-software/ff-client-sdk-go/pkg/logger"
	"reflect"
	"time"
)

type lruCache struct {
	*lru.Cache
	logger     logger.Logger
	lastUpdate time.Time
}

func NewLruCache(size int, logger logger.Logger) (*lruCache, error) {
	cache, err := lru.New(size)
	if err != nil {
		logger.Errorf("Error initializing LRU cache, err: %v", err)
		return nil, err
	}
	logger.Infof("Cache successfully initialized with size: %d", size)
	return &lruCache{
		Cache:  cache,
		logger: logger,
	}, nil
}

func (lru *lruCache) getTime() time.Time {
	return time.Now()
}

func (lru *lruCache) Set(key interface{}, value interface{}) (evicted bool) {
	prev, _ := lru.Get(key)
	if !reflect.DeepEqual(prev, value) {
		add := lru.Cache.Add(key, value)
		lru.lastUpdate = lru.getTime()
		lru.logger.Infof("cache value changed for key %s with value %v", key, value)
		return add
	}
	return false
}

func (lru *lruCache) Contains(key interface{}) bool {
	return lru.Cache.Contains(key)
}

func (lru *lruCache) ContainsOrAdd(key interface{}, value interface{}) (ok bool, evicted bool) {
	ok, evicted = lru.Cache.ContainsOrAdd(key, value)
	lru.lastUpdate = lru.getTime()
	return
}

func (lru *lruCache) Get(key interface{}) (value interface{}, ok bool) {
	return lru.Cache.Get(key)
}

func (lru *lruCache) GetOldest() (key interface{}, value interface{}, ok bool) {
	return lru.Cache.GetOldest()
}

func (lru *lruCache) Keys() []interface{} {
	return lru.Cache.Keys()
}

func (lru *lruCache) Len() int {
	return lru.Cache.Len()
}

func (lru *lruCache) Peek(key interface{}) (value interface{}, ok bool) {
	return lru.Cache.Peek(key)
}

func (lru *lruCache) PeekOrAdd(key interface{}, value interface{}) (previous interface{}, ok bool, evicted bool) {
	previous, ok, evicted = lru.Cache.PeekOrAdd(key, value)
	lru.lastUpdate = lru.getTime()
	return
}

func (lru *lruCache) Purge() {
	lru.Cache.Purge()
	lru.lastUpdate = lru.getTime()
}

func (lru *lruCache) Remove(key interface{}) (present bool) {
	present = lru.Cache.Remove(key)
	lru.lastUpdate = lru.getTime()
	if present {
		lru.logger.Infof("Cache item successfully removed %v", key)
	}
	return
}

func (lru *lruCache) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	key, value, ok = lru.Cache.RemoveOldest()
	lru.lastUpdate = lru.getTime()
	if ok {
		lru.logger.Infof("Cache oldest item successfully removed %v with value %v", key, value)
	}
	return
}

func (lru *lruCache) Resize(size int) (evicted int) {
	return lru.Cache.Resize(size)
}

func (lru *lruCache) Updated() time.Time {
	return lru.lastUpdate
}

func (lru lruCache) SetLogger(logger logger.Logger) {
	lru.logger = logger
}

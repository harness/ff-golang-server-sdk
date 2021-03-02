package cache

import (
	"github.com/drone/ff-golang-server-sdk/logger"
	"time"
)

type Cache interface {
	Set(key, value interface{}) (evicted bool)
	Contains(key interface{}) bool
	ContainsOrAdd(key, value interface{}) (ok, evicted bool)
	Get(key interface{}) (value interface{}, ok bool)
	GetOldest() (key, value interface{}, ok bool)
	Keys() []interface{}
	Len() int
	Peek(key interface{}) (value interface{}, ok bool)
	PeekOrAdd(key, value interface{}) (previous interface{}, ok, evicted bool)
	Purge()
	Remove(key interface{}) (present bool)
	RemoveOldest() (key, value interface{}, ok bool)
	Resize(size int) (evicted int)
	Updated() time.Time
	SetLogger(logger logger.Logger)
}

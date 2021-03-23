package cache

import (
	"github.com/drone/ff-golang-server-sdk.v0/logger"

	"time"
)

// Cache wrapper to integrate any 3rd party implementation
type Cache interface {
	Set(key, value interface{}) (evicted bool)
	Contains(key interface{}) bool
	Get(key interface{}) (value interface{}, ok bool)
	Keys() []interface{}
	Len() int
	Purge()
	Remove(key interface{}) (present bool)
	Resize(size int) (evicted int)
	Updated() time.Time
	SetLogger(logger logger.Logger)
}

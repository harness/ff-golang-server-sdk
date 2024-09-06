package analyticsservice

import (
	"sync"
	"sync/atomic"
)

type safeSeenTargets struct {
	sync.RWMutex
	data          map[string]bool
	maxSize       int
	limitExceeded atomic.Bool
}

// Implements SafeSeenTargetsCache
func newSafeSeenTargets(maxSize int) SafeSeenTargetsCache[string, bool] {
	return &safeSeenTargets{
		data:    make(map[string]bool),
		maxSize: maxSize,
	}
}

func (s *safeSeenTargets) setWithLimit(key string, seen bool) {
	if s.limitExceeded.Load() {
		return
	}

	s.Lock()
	defer s.Unlock()

	if len(s.data) >= s.maxSize {
		s.limitExceeded.Store(true)
		return
	}

	s.data[key] = seen
}

// The regular set method just calls SetWithLimit
func (s *safeSeenTargets) set(key string, seen bool) {
	s.setWithLimit(key, seen)
}

func (s *safeSeenTargets) get(key string) (bool, bool) {
	s.RLock()
	defer s.RUnlock()
	seen, exists := s.data[key]
	return seen, exists
}

func (s *safeSeenTargets) delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, key)
}

func (s *safeSeenTargets) size() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.data)
}

func (s *safeSeenTargets) clear() {
	s.Lock()
	defer s.Unlock()
	s.data = make(map[string]bool)
	s.limitExceeded.Store(false)
}

func (s *safeSeenTargets) iterate(f func(string, bool)) {
	s.RLock()
	defer s.RUnlock()
	for key, value := range s.data {
		f(key, value)
	}
}

func (s *safeSeenTargets) isLimitExceeded() bool {
	return s.limitExceeded.Load()
}

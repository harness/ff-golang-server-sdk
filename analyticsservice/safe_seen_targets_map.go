package analyticsservice

import (
	"maps"
	"sync"
)

type safeSeenTargets struct {
	sync.RWMutex
	data map[string]bool
}

func newSafeSeenTargets() SafeCache[string, bool] {
	return &safeSeenTargets{
		data: make(map[string]bool),
	}
}

func (s *safeSeenTargets) set(key string, seen bool) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = seen
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

func (s *safeSeenTargets) copy() SafeCache[string, bool] {
	s.RLock()
	defer s.RUnlock()
	deepCopy := make(map[string]bool)
	maps.Copy(s.data, deepCopy)
	return &safeSeenTargets{
		data: deepCopy,
	}
}

func (s *safeSeenTargets) clear() {
	s.Lock()
	defer s.Unlock()
	s.data = make(map[string]bool)
}

package analyticsservice

import "sync"

type safeSeenTargets struct {
	sync.RWMutex
	data map[string]bool
}

func newSafeSeenTargets() *safeSeenTargets {
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

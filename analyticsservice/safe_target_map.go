package analyticsservice

import (
	"sync"

	"github.com/harness/ff-golang-server-sdk/evaluation"
)

type safeTargetAnalytics struct {
	sync.RWMutex
	data map[string]evaluation.Target
}

func newSafeTargetAnalytics() SafeCache[string, evaluation.Target] {
	return &safeTargetAnalytics{
		data: make(map[string]evaluation.Target),
	}
}

func (s *safeTargetAnalytics) set(key string, value evaluation.Target) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = value
}

func (s *safeTargetAnalytics) get(key string) (evaluation.Target, bool) {
	s.RLock()
	defer s.RUnlock()
	val, exists := s.data[key]
	return val, exists
}

func (s *safeTargetAnalytics) size() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.data)
}

func (s *safeTargetAnalytics) delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, key)
}

func (s *safeTargetAnalytics) clear() {
	s.Lock()
	defer s.Unlock()
	s.data = make(map[string]evaluation.Target)
}

package analyticsservice

import (
	"maps"
	"sync"
)

type safeEvaluationAnalytics struct {
	sync.RWMutex
	data map[string]analyticsEvent
}

func newSafeEvaluationAnalytics() SafeCache[string, analyticsEvent] {
	return &safeEvaluationAnalytics{
		data: make(map[string]analyticsEvent),
	}
}

func (s *safeEvaluationAnalytics) set(key string, value analyticsEvent) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = value
}

func (s *safeEvaluationAnalytics) get(key string) (analyticsEvent, bool) {
	s.RLock()
	defer s.RUnlock()
	val, exists := s.data[key]
	return val, exists
}

func (s *safeEvaluationAnalytics) size() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.data)
}

func (s *safeEvaluationAnalytics) delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, key)
}

func (s *safeEvaluationAnalytics) copy() SafeCache[string, analyticsEvent] {
	s.RLock()
	defer s.RUnlock()
	deepCopy := make(map[string]analyticsEvent)
	maps.Copy(s.data, deepCopy)
	return &safeEvaluationAnalytics{
		data: deepCopy,
	}
}

func (s *safeEvaluationAnalytics) clear() {
	s.Lock()
	defer s.Unlock()
	s.data = make(map[string]analyticsEvent)
}

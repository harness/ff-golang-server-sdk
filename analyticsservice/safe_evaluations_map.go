package analyticsservice

import "sync"

type safeEvaluationAnalytics struct {
	sync.RWMutex
	data map[string]analyticsEvent
}

func newSafeEvaluationAnalytics() *safeEvaluationAnalytics {
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

func (s *safeEvaluationAnalytics) delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, key)
}

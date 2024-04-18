package analyticsservice

import (
	"reflect"
	"sync"
	"testing"

	"github.com/harness/ff-golang-server-sdk/evaluation"
)

// SafeMap is a generic thread-safe map
type SafeMap[K comparable, V any] struct {
	sync.RWMutex
	data map[K]V
}

// NewSafeMap creates a new SafeMap
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}

// Set sets a value in the map
func (s *SafeMap[K, V]) Set(key K, value V) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = value
}

// Get retrieves a value from the map
func (s *SafeMap[K, V]) Get(key K) (V, bool) {
	s.RLock()
	defer s.RUnlock()
	val, exists := s.data[key]
	return val, exists
}

// Delete removes a value from the map
func (s *SafeMap[K, V]) Delete(key K) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, key)
}

func testSafeMapOperations[K comparable, V any](t *testing.T, mapInstance *SafeMap[K, V], testData map[K]V) {
	// Test set and get
	for key, value := range testData {
		mapInstance.Set(key, value)
		if got, exists := mapInstance.Get(key); !exists || !reflect.DeepEqual(got, value) {
			t.Errorf("set or get method failed for key %v, expected %v, got %v", key, value, got)
		}
	}

	// Test concurrent access
	var wg sync.WaitGroup
	for key, value := range testData {
		wg.Add(1)
		go func(k K, v V) {
			defer wg.Done()
			mapInstance.Set(k, v)
			if got, exists := mapInstance.Get(k); !exists || !reflect.DeepEqual(got, v) {
				t.Errorf("concurrent set or get failed for key %v, expected %v, got %v", k, v, got)
			}
		}(key, value)
	}
	wg.Wait()

	// Test delete
	for key := range testData {
		mapInstance.Delete(key)
		if _, exists := mapInstance.Get(key); exists {
			t.Errorf("delete method failed, %v should have been deleted", key)
		}
	}
}

func TestSafeEvaluationAnalytics(t *testing.T) {
	EvaluationAnalytics := NewSafeMap[string, analyticsEvent]()
	testData := map[string]analyticsEvent{
		"test-key": {count: 1},
		"key-1":    {count: 10},
		"key-2":    {count: 20},
	}
	testSafeMapOperations(t, EvaluationAnalytics, testData)
}

func TestSafeTargetAnalytics(t *testing.T) {
	TargetAnalytics := NewSafeMap[string, evaluation.Target]()
	testData := map[string]evaluation.Target{
		"test-target": {Identifier: "test-identifier"},
		"target-1":    {Identifier: "id-10"},
		"target-2":    {Identifier: "id-20"},
	}
	testSafeMapOperations(t, TargetAnalytics, testData)
}

func TestSafeSeenTargets(t *testing.T) {
	SeenTargets := NewSafeMap[string, bool]()

	testData := map[string]bool{
		"test-seen": true,
		"seen-1":    false,
		"seen-2":    true,
	}
	testSafeMapOperations(t, SeenTargets, testData)
}

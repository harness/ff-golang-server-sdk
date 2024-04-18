package analyticsservice

import (
	"reflect"
	"sync"
	"testing"

	"github.com/harness/ff-golang-server-sdk/evaluation"
)

func testMapOperations[K comparable, V any](t *testing.T, mapInstance MapOperations[K, V], testData map[K]V) {
	var wg sync.WaitGroup

	// Test concurrent sets and gets
	for key, value := range testData {
		wg.Add(1)
		go func(k K, v V) {
			defer wg.Done()
			mapInstance.set(k, v)
			if got, exists := mapInstance.get(k); !exists || !reflect.DeepEqual(got, v) {
				t.Errorf("Concurrent set or get method failed for key %v, expected %v, got %v", k, v, got)
			}
		}(key, value)
	}
	wg.Wait()

	// Test concurrent deletes
	for key := range testData {
		wg.Add(1)
		go func(k K) {
			defer wg.Done()
			mapInstance.delete(k)
			if _, exists := mapInstance.get(k); exists {
				t.Errorf("Concurrent delete method failed, %v should have been deleted", k)
			}
		}(key)
	}
	wg.Wait() // Wait for all delete operations to complete
}

func TestSafeEvaluationAnalytics(t *testing.T) {
	s := newSafeEvaluationAnalytics()
	testData := map[string]analyticsEvent{
		"event1": {count: 10},
		"event2": {count: 5},
		"event3": {count: 5},
		"event4": {count: 3},
		"event5": {count: 2},
		"event6": {count: 1},
	}

	testMapOperations[string, analyticsEvent](t, s, testData)
}

func TestSafeTargetAnalytics(t *testing.T) {
	s := newSafeTargetAnalytics()
	testData := map[string]evaluation.Target{
		"target1": {Identifier: "id1"},
		"target2": {Identifier: "id2"},
		"target3": {Identifier: "id3"},
		"target4": {Identifier: "id4"},
		"target5": {Identifier: "id5"},
	}

	testMapOperations[string, evaluation.Target](t, s, testData)
}

func TestSafeSeenTargets(t *testing.T) {
	s := newSafeSeenTargets()
	testData := map[string]bool{
		"target1":  true,
		"target21": true,
		"target3":  true,
		"target4":  true,
	}

	testMapOperations[string, bool](t, s, testData)
}

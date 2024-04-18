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

//
//import (
//	"reflect"
//	"sync"
//	"testing"
//
//	"github.com/harness/ff-golang-server-sdk/evaluation"
//)
//
//func testSafeMapOperations[K comparable, V any](t *testing.T, testData map[K]V, setFunc func(K, V), getFunc func(K) (V, bool), deleteFunc func(K)) {
//	for key, value := range testData {
//		setFunc(key, value)
//		if got, exists := getFunc(key); !exists || !reflect.DeepEqual(got, value) {
//			t.Errorf("set or get method failed for key %v, expected %v, got %v", key, value, got)
//		}
//	}
//
//	var wg sync.WaitGroup
//	for key, value := range testData {
//		wg.Add(1)
//		go func(k K, v V) {
//			defer wg.Done()
//			setFunc(k, v)
//			if got, exists := getFunc(k); !exists || !reflect.DeepEqual(got, v) {
//				t.Errorf("Concurrent set or get failed for key %v, expected %v, got %v", k, v, got)
//			}
//		}(key, value)
//	}
//	wg.Wait()
//
//	for key := range testData {
//		deleteFunc(key)
//		if _, exists := getFunc(key); exists {
//			t.Errorf("delete method failed, %v should have been deleted", key)
//		}
//	}
//}
//
//func TestSafeTargetAnalytics(t *testing.T) {
//	s := newSafeTargetAnalytics()
//	testData := map[string]evaluation.Target{
//		"target1": {Identifier: "id1"},
//		"target2": {Identifier: "id2"},
//	}
//
//	testSafeMapOperations(t, testData,
//		func(key string, value evaluation.Target) { s.set(key, value) },
//		func(key string) (evaluation.Target, bool) { return s.get(key) },
//		func(key string) { s.delete(key) },
//	)
//}
//
//func TestSafeEvaluationAnalytics(t *testing.T) {
//	s := newSafeEvaluationAnalytics()
//	testData := map[string]analyticsEvent{
//		"event1": {count: 1},
//		"event2": {count: 2},
//	}
//
//	testSafeMapOperations(t, testData,
//		func(key string, value analyticsEvent) { s.set(key, value) },
//		func(key string) (analyticsEvent, bool) { return s.get(key) },
//		func(key string) { s.delete(key) },
//	)
//}
//
//func TestSafeSeenTargets(t *testing.T) {
//	s := newSafeSeenTargets()
//	testData := map[string]bool{
//		"seen1": true,
//		"seen2": false,
//	}
//
//	testSafeMapOperations(t, testData,
//		func(key string, value bool) { s.set(key, value) },
//		func(key string) (bool, bool) { return s.get(key) },
//		func(key string) { s.delete(key) },
//	)
//}

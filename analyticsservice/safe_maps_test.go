package analyticsservice

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/harness/ff-golang-server-sdk/evaluation"
)

func testMapOperations[K comparable, V any](t *testing.T, mapInstance SafeAnalyticsCache[K, V], testData map[K]V) {
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

	// Test concurrent iteration and size
	for key := range testData {
		wg.Add(1)
		go func(k K) {
			defer wg.Done()
			mapInstance.size()
			mapInstance.iterate(func(k K, v V) {
				if expected, exists := testData[k]; !exists || !reflect.DeepEqual(v, expected) {
					t.Errorf("Iterate failed for key %v, expected %v, got %v", k, expected, v)
				}
			})
		}(key)
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
	wg.Wait()
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
	// Initialize with a small maxSize for testing
	maxSize := 3
	s := newSafeSeenTargets(maxSize, 0).(SafeSeenTargetsCache[string, bool])

	testData := map[string]bool{
		"target1":  true,
		"target21": true,
		"target3":  true,
	}

	// Insert items and ensure limit is not exceeded
	for key, value := range testData {
		s.set(key, value)
	}

	if s.isLimitExceeded() {
		t.Errorf("Limit should not have been exceeded yet")
	}

	// Add one more item to exceed the limit
	s.setWithLimit("target4", true)

	// Ensure limitExceeded is true after exceeding the limit
	if !s.isLimitExceeded() {
		t.Errorf("Limit should be exceeded after adding target4")
	}

	// Ensure that new items are not added once the limit is exceeded
	s.setWithLimit("target5", true)
	if _, exists := s.get("target5"); exists {
		t.Errorf("target5 should not have been added as the limit was exceeded")
	}

	// Clear the map and ensure limit is reset
	s.clear()

	if s.isLimitExceeded() {
		t.Errorf("Limit should have been reset after clearing the map")
	}

	// Add items again after clearing
	s.setWithLimit("target6", true)
	if _, exists := s.get("target6"); !exists {
		t.Errorf("target6 should have been added after clearing the map")
	}

	// Concurrency test
	t.Run("ConcurrencyTest", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrencyLevel := 100

		// Re-initialize the map for concurrency testing
		s = newSafeSeenTargets(100, 0).(SafeSeenTargetsCache[string, bool])

		// Concurrently set keys
		for i := 0; i < concurrencyLevel; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := "target" + fmt.Sprint(i)
				s.setWithLimit(key, true)
			}(i)
		}

		// Concurrently get keys
		for i := 0; i < concurrencyLevel; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := "target" + fmt.Sprint(i)
				s.get(key)
			}(i)
		}

		// Concurrently clear the map
		for i := 0; i < concurrencyLevel/2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.clear()
			}()
		}

		wg.Wait()

		// Ensure the map is cleared after the concurrency operations
		if s.size() > 0 {
			t.Errorf("Map size should be 0 after clearing, got %d", s.size())
		}
	})

	// Add test for clearing based on interval
	t.Run("IntervalClearingTest", func(t *testing.T) {
		// Re-initialize the map with a clearing interval
		s = newSafeSeenTargets(10, 100*time.Millisecond)

		for key, value := range testData {
			s.set(key, value)
		}

		// Ensure the map has items initially
		if s.size() != len(testData) {
			t.Errorf("Expected map size to be %d, got %d", len(testData), s.size())
		}

		// Wait for the clearing to clear the map
		time.Sleep(300 * time.Millisecond)

		// Ensure the map is cleared after the interval
		if s.size() != 0 {
			t.Errorf("Expected map size to be 0 after clearing interval, got %d", s.size())
		}

		// Ensure the limitExceeded flag is reset
		if s.isLimitExceeded() {
			t.Errorf("Expected limitExceeded to be reset after clearing")
		}
	})
}

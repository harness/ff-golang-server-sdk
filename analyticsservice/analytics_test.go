package analyticsservice

import (
	"testing"
	"time"

	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/rest"
)

func TestListenerHandlesEventsCorrectly(t *testing.T) {
	noOpLogger := logger.NewNoOpLogger()

	testCases := []struct {
		name                string
		events              []analyticsEvent
		expectedEvaluations map[string]int
		expectedSeen        map[string]bool
		expectedTargets     map[string]evaluation.Target
	}{
		{
			name: "Single evaluation",
			events: []analyticsEvent{
				{target: &evaluation.Target{Identifier: "target1"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
			},
			expectedEvaluations: map[string]int{"feature1-var1-value1-global": 1},
			expectedSeen:        map[string]bool{"target1": true},
			expectedTargets:     map[string]evaluation.Target{"target1": {Identifier: "target1"}},
		},
		{
			name: "Two identical evaluations with the same target",
			events: []analyticsEvent{
				{target: &evaluation.Target{Identifier: "target1"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
				{target: &evaluation.Target{Identifier: "target1"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
			},
			expectedEvaluations: map[string]int{"feature1-var1-value1-global": 2},
			expectedSeen:        map[string]bool{"target1": true},
			expectedTargets:     map[string]evaluation.Target{"target1": {Identifier: "target1"}},
		},
		{
			name: "Two identical evaluations with different targets",
			events: []analyticsEvent{
				{target: &evaluation.Target{Identifier: "target1"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
				{target: &evaluation.Target{Identifier: "target2"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
			},
			expectedEvaluations: map[string]int{"feature1-var1-value1-global": 2},
			expectedSeen:        map[string]bool{"target1": true, "target2": true},
			expectedTargets:     map[string]evaluation.Target{"target1": {Identifier: "target1"}, "target2": {Identifier: "target2"}},
		},
		{
			name: "Two different evaluations with two different targets",
			events: []analyticsEvent{
				{target: &evaluation.Target{Identifier: "target1"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
				{target: &evaluation.Target{Identifier: "target2"}, featureConfig: &rest.FeatureConfig{Feature: "feature2"}, variation: &rest.Variation{Identifier: "var2", Value: "value2"}},
			},
			expectedEvaluations: map[string]int{"feature1-var1-value1-global": 1, "feature2-var2-value2-global": 1},
			expectedSeen:        map[string]bool{"target1": true, "target2": true},
			expectedTargets:     map[string]evaluation.Target{"target1": {Identifier: "target1"}, "target2": {Identifier: "target2"}},
		},
		{
			name: "Three different evaluations with two identical targets",
			events: []analyticsEvent{
				{target: &evaluation.Target{Identifier: "target1"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
				{target: &evaluation.Target{Identifier: "target2"}, featureConfig: &rest.FeatureConfig{Feature: "feature2"}, variation: &rest.Variation{Identifier: "var2", Value: "value2"}},
				{target: &evaluation.Target{Identifier: "target3"}, featureConfig: &rest.FeatureConfig{Feature: "feature3"}, variation: &rest.Variation{Identifier: "var3", Value: "value3"}},
			},
			expectedEvaluations: map[string]int{"feature1-var1-value1-global": 1, "feature2-var2-value2-global": 1, "feature3-var3-value3-global": 1},
			expectedSeen:        map[string]bool{"target1": true, "target2": true, "target3": true},
			expectedTargets:     map[string]evaluation.Target{"target1": {Identifier: "target1"}, "target2": {Identifier: "target2"}, "target3": {Identifier: "target3"}},
		},
		{
			name: "Three different evaluations with two anonymous targets",
			events: []analyticsEvent{
				{target: &evaluation.Target{Identifier: "target1"}, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
				{target: &evaluation.Target{Identifier: "target2", Anonymous: boolPtr(true)}, featureConfig: &rest.FeatureConfig{Feature: "feature2"}, variation: &rest.Variation{Identifier: "var2", Value: "value2"}},
				{target: &evaluation.Target{Identifier: "target3"}, featureConfig: &rest.FeatureConfig{Feature: "feature3"}, variation: &rest.Variation{Identifier: "var3", Value: "value3"}},
			},
			expectedEvaluations: map[string]int{"feature1-var1-value1-global": 1, "feature2-var2-value2-global": 1, "feature3-var3-value3-global": 1},
			expectedSeen:        map[string]bool{"target1": true, "target3": true},
			expectedTargets:     map[string]evaluation.Target{"target1": {Identifier: "target1"}, "target3": {Identifier: "target3"}},
		},
		{
			name: "Three different evaluations with one anonymous target and one nil target",
			events: []analyticsEvent{
				{target: nil, featureConfig: &rest.FeatureConfig{Feature: "feature1"}, variation: &rest.Variation{Identifier: "var1", Value: "value1"}},
				{target: &evaluation.Target{Identifier: "target2", Anonymous: boolPtr(true)}, featureConfig: &rest.FeatureConfig{Feature: "feature2"}, variation: &rest.Variation{Identifier: "var2", Value: "value2"}},
				{target: &evaluation.Target{Identifier: "target3"}, featureConfig: &rest.FeatureConfig{Feature: "feature3"}, variation: &rest.Variation{Identifier: "var3", Value: "value3"}},
			},
			expectedEvaluations: map[string]int{"feature1-var1-value1-global": 1, "feature2-var2-value2-global": 1, "feature3-var3-value3-global": 1},
			expectedSeen:        map[string]bool{"target3": true},
			expectedTargets:     map[string]evaluation.Target{"target3": {Identifier: "target3"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewAnalyticsService(1*time.Minute, noOpLogger)
			defer close(service.analyticsChan)

			// Start the listener in a goroutine
			go service.listener()

			// Send all events for the test case
			for _, event := range tc.events {
				service.analyticsChan <- event
			}

			// Allow some time for the events to be processed
			time.Sleep(100 * time.Millisecond)

			// Check evaluation metrics counts
			for key, expectedCount := range tc.expectedEvaluations {
				analytic, exists := service.evaluationAnalytics.get(key)
				if !exists || analytic.count != expectedCount {
					t.Errorf("Test %s failed: expected count for key %s is %d, got %d", tc.name, key, expectedCount, analytic.count)
				}
			}

			// Check target metrics
			for targetID, expectedSeen := range tc.expectedSeen {
				if _, seen := service.seenTargets.get(targetID); seen != expectedSeen {
					t.Errorf("Test %s failed: expected target to be in seen targets cache %s is %v", tc.name, targetID, expectedSeen)
				}
			}

			// Check target analytics
			for targetID, expectedTarget := range tc.expectedTargets {
				target, exists := service.targetAnalytics.get(targetID)
				if !exists || target.Identifier != expectedTarget.Identifier {
					t.Errorf("Test %s failed: expected target to be in target cache %s", tc.name, targetID)
				}
			}
		})
	}
}

func Test_convertInterfaceToString(t *testing.T) {
	testCases := map[string]struct {
		input    interface{}
		expected string
	}{
		"Given I input a string": {
			input:    "123",
			expected: "123",
		},
		"Given I input a bool with the value false": {
			input:    false,
			expected: "false",
		},
		"Given I input a bool with the value true": {
			input:    true,
			expected: "true",
		},
		"Given I input an int64": {
			input:    int64(123),
			expected: "123",
		},
		"Given I input an int": {
			input:    123,
			expected: "123",
		},
		"Given I input a float64": {
			input:    float64(2.5),
			expected: "2.5",
		},
		"Given I input a float32": {
			input:    float32(2.5),
			expected: "2.5",
		},
		"Given I input a nil value": {
			input:    nil,
			expected: "nil",
		},
	}

	for desc, tc := range testCases {
		tc := tc
		t.Run(desc, func(t *testing.T) {

			actual := convertInterfaceToString(tc.input)
			if actual != tc.expected {
				t.Errorf("(%s): expected %s, actual %s", desc, tc.expected, actual)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

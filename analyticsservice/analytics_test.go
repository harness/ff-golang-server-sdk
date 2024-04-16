package analyticsservice

import (
	"testing"
	"time"

	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/rest"
)

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

func TestListenerHandlesEventsCorrectly(t *testing.T) {
	noOpLogger := logger.NewNoOpLogger() // assuming you have a constructor for a noOpLogger
	service := NewAnalyticsService(1*time.Minute, noOpLogger)
	defer close(service.analyticsChan) // Ensure the channel is closed after test

	target := &evaluation.Target{Identifier: "target1", Anonymous: new(bool)}
	featureConfig := &rest.FeatureConfig{Feature: "feature1"}
	variation := &rest.Variation{Identifier: "var1", Value: "value1"}

	// Send an event to the channel
	go func() {
		service.analyticsChan <- analyticsEvent{
			target:        target,
			featureConfig: featureConfig,
			variation:     variation,
		}
	}()

	// Allow some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	// Check if the event is processed correctly
	service.evaluationsAnalyticsMx.Lock()
	if len(service.evaluationAnalytics) != 1 {
		t.Errorf("Expected evaluationAnalytics to contain 1 item, got %d", len(service.evaluationAnalytics))
	}
	service.evaluationsAnalyticsMx.Unlock()

	service.seenTargetsMx.RLock()
	if !service.seenTargets["target1"] {
		t.Errorf("Expected target 'target1' to be marked as seen")
	}
	service.seenTargetsMx.RUnlock()
}

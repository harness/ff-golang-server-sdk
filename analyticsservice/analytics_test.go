package analyticsservice

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/metricsclient"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsClient is a mock implementation for the ClientWithResponsesInterface
type MockMetricsClient struct {
	mock.Mock
}

func (m *MockMetricsClient) PostMetricsWithResponse(ctx context.Context, environmentUUID metricsclient.EnvironmentPathParam, params *metricsclient.PostMetricsParams, body metricsclient.PostMetricsJSONRequestBody, reqEditors ...metricsclient.RequestEditorFn) (*metricsclient.PostMetricsResponse, error) {
	args := m.Called(ctx, environmentUUID, params, body)
	return args.Get(0).(*metricsclient.PostMetricsResponse), args.Error(1)
}

func (m *MockMetricsClient) PostMetricsWithBodyWithResponse(ctx context.Context, environmentUUID metricsclient.EnvironmentPathParam, params *metricsclient.PostMetricsParams, contentType string, body io.Reader, reqEditors ...metricsclient.RequestEditorFn) (*metricsclient.PostMetricsResponse, error) {
	args := m.Called(ctx, environmentUUID, params, contentType, body)
	return args.Get(0).(*metricsclient.PostMetricsResponse), args.Error(1)
}

func TestSendDataAndResetCache(t *testing.T) {
	noOpLogger := logger.NewNoOpLogger()
	mockMetricsClient := MockMetricsClient{}
	service := NewAnalyticsService(1*time.Minute, noOpLogger)
	service.metricsClient = &mockMetricsClient

	// Setup mock response from metrics server
	mockResponse := &metricsclient.PostMetricsResponse{}
	mockMetricsClient.On("PostMetricsWithResponse", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Simulate analytics data
	service.evaluationAnalytics["test-key"] = analyticsEvent{
		target:        &evaluation.Target{Identifier: "target1", Name: "Target One"},
		featureConfig: &rest.FeatureConfig{Feature: "feature1"},
		variation:     &rest.Variation{Identifier: "var1", Value: "value1"},
		count:         1,
	}
	service.targetAnalytics["target1"] = evaluation.Target{Identifier: "target1", Name: "Target One", Attributes: &map[string]interface{}{"key": "value"}}

	ctx := context.TODO()
	service.sendDataAndResetCache(ctx)

	// Verify that metrics data is sent correctly
	mockMetricsClient.AssertCalled(t, "PostMetricsWithResponse", ctx, service.environmentID, nil, mock.AnythingOfType("metricsclient.PostMetricsJSONRequestBody"))

	// Verify that caches are reset
	assert.Empty(t, service.evaluationAnalytics, "Evaluation analytics cache should be empty after sending data")
	assert.Empty(t, service.targetAnalytics, "Target analytics cache should be empty after sending data")
}

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
			service.evaluationsAnalyticsMx.Lock()
			for key, expectedCount := range tc.expectedEvaluations {
				analytic, exists := service.evaluationAnalytics[key]
				if !exists || analytic.count != expectedCount {
					t.Errorf("Test %s failed: expected count for key %s is %d, got %d", tc.name, key, expectedCount, analytic.count)
				}
			}
			service.evaluationsAnalyticsMx.Unlock()

			// Check target metrics
			service.seenTargetsMx.RLock()
			for targetID, expectedSeen := range tc.expectedSeen {
				if seen := service.seenTargets[targetID]; seen != expectedSeen {
					t.Errorf("Test %s failed: expected target to be in seen targets cache %s is %v", tc.name, targetID, expectedSeen)
				}
			}
			service.seenTargetsMx.RUnlock()

			// Check target analytics
			service.targetAnalyticsMx.Lock()
			for targetID, expectedTarget := range tc.expectedTargets {
				target, exists := service.targetAnalytics[targetID]
				if !exists || target.Identifier != expectedTarget.Identifier {
					t.Errorf("Test %s failed: expected target to be in target cache %s", tc.name, targetID)
				}
			}
			service.targetAnalyticsMx.Unlock()
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

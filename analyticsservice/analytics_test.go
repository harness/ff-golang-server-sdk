package analyticsservice

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/metricsclient"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/stretchr/testify/assert"
)

type MockMetricsClient struct {
	LastBody  metricsclient.PostMetricsJSONRequestBody
	CallCount int
}

func (c *MockMetricsClient) PostMetricsWithResponse(ctx context.Context, environmentUUID metricsclient.EnvironmentPathParam, params *metricsclient.PostMetricsParams, body metricsclient.PostMetricsJSONRequestBody, reqEditors ...metricsclient.RequestEditorFn) (*metricsclient.PostMetricsResponse, error) {
	c.LastBody = body
	c.CallCount++

	return &metricsclient.PostMetricsResponse{
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}, nil
}

func (c *MockMetricsClient) PostMetricsWithBodyWithResponse(ctx context.Context, environmentUUID metricsclient.EnvironmentPathParam, params *metricsclient.PostMetricsParams, contentType string, body io.Reader, reqEditors ...metricsclient.RequestEditorFn) (*metricsclient.PostMetricsResponse, error) {

	return nil, nil // default to returning no error and a nil response
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

func Test_ProcessEvaluationMetrics(t *testing.T) {
	var timeStamp int64 = 1715600410545
	testCases := []struct {
		name      string
		events    map[string]analyticsEvent
		expected  []metricsclient.MetricsData
		expLength int
	}{
		{
			name: "One unique evaluation evaluated 10 times",
			events: map[string]analyticsEvent{
				"key1": {
					featureConfig: &rest.FeatureConfig{Feature: "feature1"},
					variation:     &rest.Variation{Identifier: "var1", Value: "value1"},
					count:         10,
				},
			},
			expLength: 1,
			expected: []metricsclient.MetricsData{
				{
					Count: 10,
					Attributes: []metricsclient.KeyValue{
						{Key: featureIdentifierAttribute, Value: "feature1"},
						{Key: featureNameAttribute, Value: "feature1"},
						{Key: variationIdentifierAttribute, Value: "var1"},
						{Key: variationValueAttribute, Value: "value1"},
						{Key: sdkTypeAttribute, Value: sdkType},
						{Key: sdkLanguageAttribute, Value: sdkLanguage},
						{Key: sdkVersionAttribute, Value: SdkVersion},
						{Key: targetAttribute, Value: globalTarget},
					},
					MetricsType: metricsclient.MetricsDataMetricsType(ffMetricType),
					Timestamp:   timeStamp,
				},
			},
		},
		{
			name: "Two unique evaluation evaluated 5 and 7 times",
			events: map[string]analyticsEvent{
				"key1": {
					featureConfig: &rest.FeatureConfig{Feature: "feature1"},
					variation:     &rest.Variation{Identifier: "var1", Value: "value1"},
					count:         5,
				},
				"key2": {
					featureConfig: &rest.FeatureConfig{Feature: "feature2"},
					variation:     &rest.Variation{Identifier: "var2", Value: "value2"},
					count:         7,
				},
			},
			expLength: 2,
			expected: []metricsclient.MetricsData{
				{
					Count: 5,
					Attributes: []metricsclient.KeyValue{
						{Key: featureIdentifierAttribute, Value: "feature1"},
						{Key: featureNameAttribute, Value: "feature1"},
						{Key: variationIdentifierAttribute, Value: "var1"},
						{Key: variationValueAttribute, Value: "value1"},
						{Key: sdkTypeAttribute, Value: sdkType},
						{Key: sdkLanguageAttribute, Value: sdkLanguage},
						{Key: sdkVersionAttribute, Value: SdkVersion},
						{Key: targetAttribute, Value: globalTarget},
					},
					MetricsType: metricsclient.MetricsDataMetricsType(ffMetricType),
					Timestamp:   timeStamp,
				},
				{
					Count: 7,
					Attributes: []metricsclient.KeyValue{
						{Key: featureIdentifierAttribute, Value: "feature2"},
						{Key: featureNameAttribute, Value: "feature2"},
						{Key: variationIdentifierAttribute, Value: "var2"},
						{Key: variationValueAttribute, Value: "value2"},
						{Key: sdkTypeAttribute, Value: sdkType},
						{Key: sdkLanguageAttribute, Value: sdkLanguage},
						{Key: sdkVersionAttribute, Value: SdkVersion},
						{Key: targetAttribute, Value: globalTarget},
					},
					MetricsType: metricsclient.MetricsDataMetricsType(ffMetricType),
					Timestamp:   timeStamp,
				},
			},
		},
		{
			name:      "No metrics",
			events:    map[string]analyticsEvent{},
			expLength: 0,
			expected:  []metricsclient.MetricsData{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := newSafeEvaluationAnalytics()
			for key, event := range tc.events {
				cache.set(key, event)
			}

			service := AnalyticsService{
				evaluationAnalytics: cache,
			}

			metrics := service.processEvaluationMetrics(cache, timeStamp)

			assert.ElementsMatch(t, tc.expected, metrics)

			if len(metrics) != tc.expLength {
				t.Errorf("Expected %d metrics data, got %d", tc.expLength, len(metrics))
			}

		})
	}
}

func Test_ProcessTargetMetrics(t *testing.T) {
	cases := []struct {
		name      string
		targets   map[string]evaluation.Target
		expLength int
		expected  []metricsclient.TargetData
	}{
		{
			name: "Single target",
			targets: map[string]evaluation.Target{
				"target1": {
					Identifier: "target1",
					Name:       "Target One",
					Attributes: &map[string]interface{}{"key": "value"},
				},
			},
			expLength: 1,
			expected: []metricsclient.TargetData{
				{
					Identifier: "target1",
					Name:       "Target One",
					Attributes: []metricsclient.KeyValue{{Key: "key", Value: "value"}},
				},
			},
		},
		{
			name: "Single target with no attributes",
			targets: map[string]evaluation.Target{
				"target1": {
					Identifier: "target1",
					Name:       "Target One",
				},
			},
			expLength: 1,
			expected: []metricsclient.TargetData{
				{
					Identifier: "target1",
					Name:       "Target One",
					Attributes: []metricsclient.KeyValue{},
				},
			},
		},
		{
			name: "Two targets, one with no attributes",
			targets: map[string]evaluation.Target{
				"target1": {
					Identifier: "target1",
					Name:       "Target One",
					Attributes: &map[string]interface{}{"key": "value"},
				},
				"target2": {
					Identifier: "target2",
					Name:       "Target Two",
				},
			},
			expLength: 2,
			expected: []metricsclient.TargetData{
				{
					Identifier: "target1",
					Name:       "Target One",
					Attributes: []metricsclient.KeyValue{{Key: "key", Value: "value"}},
				},
				{
					Identifier: "target2",
					Name:       "Target Two",
					Attributes: []metricsclient.KeyValue{},
				},
			},
		},
		{
			name:      "No targets",
			targets:   map[string]evaluation.Target{},
			expLength: 0,
			expected:  []metricsclient.TargetData{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cache := newSafeTargetAnalytics()
			for key, target := range tc.targets {
				cache.set(key, target)
			}

			service := AnalyticsService{
				targetAnalytics: cache,
			}

			targetMetrics := service.processTargetMetrics(cache)
			if len(targetMetrics) != tc.expLength {
				t.Errorf("Expected %d target metrics, got %d", tc.expLength, len(targetMetrics))
			}

			assert.ElementsMatch(t, tc.expected, targetMetrics)
		})
	}
}

func TestSendDataAndResetCache(t *testing.T) {
	noOpLogger := logger.NewNoOpLogger()

	evaluationAnalytics := newSafeEvaluationAnalytics()
	evaluationAnalytics.set("key1", analyticsEvent{
		featureConfig: &rest.FeatureConfig{Feature: "feature1"},
		variation:     &rest.Variation{Identifier: "var1", Value: "value1"},
		count:         3,
	})
	evaluationAnalytics.set("key2", analyticsEvent{
		featureConfig: &rest.FeatureConfig{Feature: "feature2"},
		variation:     &rest.Variation{Identifier: "var2", Value: "value2"},
		count:         1,
	})

	targetAnalytics := newSafeTargetAnalytics()

	// Use anonymous and attributes
	targetAnalytics.set("target1", evaluation.Target{
		Identifier: "target1",
		Name:       "Target One",
		Anonymous:  boolPtr(false),
		Attributes: &map[string]interface{}{"key1": "value1"},
	})

	// No attributes or anonymous set
	targetAnalytics.set("target2", evaluation.Target{
		Identifier: "target2",
		Name:       "Target Two",
	})

	mClient := &MockMetricsClient{}

	service := AnalyticsService{
		evaluationAnalytics: evaluationAnalytics,
		targetAnalytics:     targetAnalytics,
		logger:              noOpLogger,
		metricsClient:       mClient,
		environmentID:       "test-env",
	}

	ctx := context.Background()

	var timeStamp int64 = 1715600410545
	service.sendDataAndResetCache(ctx, timeStamp)

	expectedAnalyticsPayload := metricsclient.PostMetricsJSONRequestBody{
		MetricsData: &[]metricsclient.MetricsData{
			{
				Count: 3,
				Attributes: []metricsclient.KeyValue{
					{Key: featureIdentifierAttribute, Value: "feature1"},
					{Key: featureNameAttribute, Value: "feature1"},
					{Key: variationIdentifierAttribute, Value: "var1"},
					{Key: variationValueAttribute, Value: "value1"},
					{Key: sdkTypeAttribute, Value: sdkType},
					{Key: sdkLanguageAttribute, Value: sdkLanguage},
					{Key: sdkVersionAttribute, Value: SdkVersion},
					{Key: targetAttribute, Value: globalTarget},
				},
				MetricsType: metricsclient.MetricsDataMetricsType(ffMetricType),
				Timestamp:   timeStamp,
			},
			{
				Count: 1,
				Attributes: []metricsclient.KeyValue{
					{Key: featureIdentifierAttribute, Value: "feature2"},
					{Key: featureNameAttribute, Value: "feature2"},
					{Key: variationIdentifierAttribute, Value: "var2"},
					{Key: variationValueAttribute, Value: "value2"},
					{Key: sdkTypeAttribute, Value: sdkType},
					{Key: sdkLanguageAttribute, Value: sdkLanguage},
					{Key: sdkVersionAttribute, Value: SdkVersion},
					{Key: targetAttribute, Value: globalTarget},
				},
				MetricsType: metricsclient.MetricsDataMetricsType(ffMetricType),
				Timestamp:   timeStamp,
			},
		},
		TargetData: &[]metricsclient.TargetData{
			{
				Identifier: "target1",
				Name:       "Target One",
				Attributes: []metricsclient.KeyValue{{Key: "key1", Value: "value1"}},
			},
			{
				Identifier: "target2",
				Name:       "Target Two",
				Attributes: []metricsclient.KeyValue{},
			},
		},
	}

	// Verify that PostMetricsWithResponse was called correctly
	assert.Equal(t, 1, mClient.CallCount, "PostMetricsWithResponse should be called once")
	assert.ElementsMatch(t, *mClient.LastBody.MetricsData, *expectedAnalyticsPayload.MetricsData)
	assert.ElementsMatch(t, *mClient.LastBody.TargetData, *expectedAnalyticsPayload.TargetData)

	// Verify both caches were reset
	if service.evaluationAnalytics.size() != 0 {
		t.Errorf("Expected the evaluation analytics cache to be reset")
	}

	if service.targetAnalytics.size() != 0 {
		t.Errorf("Expected the target analytics cache to be reset")
	}
}

func boolPtr(b bool) *bool {
	return &b
}

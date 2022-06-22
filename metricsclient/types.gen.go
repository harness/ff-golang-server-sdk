// Package metricsclient provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package metricsclient

const (
	BearerAuthScopes = "BearerAuth.Scopes"
)

// Defines values for MetricsDataMetricsType.
const (
	FFMETRICS MetricsDataMetricsType = "FFMETRICS"
)

// Error defines model for Error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// KeyValue defines model for KeyValue.
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Metrics defines model for Metrics.
type Metrics struct {
	MetricsData *[]MetricsData `json:"metricsData,omitempty"`
	TargetData  *[]TargetData  `json:"targetData,omitempty"`
}

// MetricsData defines model for MetricsData.
type MetricsData struct {
	Attributes []KeyValue `json:"attributes"`
	Count      int        `json:"count"`

	// This can be of type FeatureMetrics
	MetricsType MetricsDataMetricsType `json:"metricsType"`

	// time at when this data was recorded
	Timestamp int64 `json:"timestamp"`
}

// This can be of type FeatureMetrics
type MetricsDataMetricsType string

// TargetData defines model for TargetData.
type TargetData struct {
	Attributes []KeyValue `json:"attributes"`
	Identifier string     `json:"identifier"`
	Name       string     `json:"name"`
}

// EnvironmentPathParam defines model for environmentPathParam.
type EnvironmentPathParam = string

// InternalServerError defines model for InternalServerError.
type InternalServerError = Error

// Unauthenticated defines model for Unauthenticated.
type Unauthenticated = Error

// Unauthorized defines model for Unauthorized.
type Unauthorized = Error

// PostMetricsJSONBody defines parameters for PostMetrics.
type PostMetricsJSONBody = Metrics

// PostMetricsJSONRequestBody defines body for PostMetrics for application/json ContentType.
type PostMetricsJSONRequestBody = PostMetricsJSONBody

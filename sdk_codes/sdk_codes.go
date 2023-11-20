package sdk_codes

// SDKCode is the type for codes that are logged for each lifecycle of the SDK
type SDKCode string

const (
	InitSuccess         SDKCode = "SDKCODE:1000"
	InitAuthError       SDKCode = "SDKCODE:1001"
	InitMissingKey      SDKCode = "SDKCODE:1002"
	InitWaiting         SDKCode = "SDKCODE:1003"
	AuthSuccess         SDKCode = "SDKCODE:2000"
	AuthFailed          SDKCode = "SDKCODE:2001"
	AuthAttempt         SDKCode = "SDKCODE:2002"
	AuthExceededRetries SDKCode = "SDKCODE:2003"
	CloseStarted        SDKCode = "SDKCODE:3000"
	CloseSuccess        SDKCode = "SDKCODE:3001"
	PollStart           SDKCode = "SDKCODE:4000"
	PollStop            SDKCode = "SDKCODE:4001"
	StreamStarted       SDKCode = "SDKCODE:5000"
	StreamDisconnected  SDKCode = "SDKCODE:5001"
	StreamEvent         SDKCode = "SDKCODE:5002"
	// StreamRetry TODO it's not clear how the SSE retry mechanism is working. Add this once SSE resilency has been established in FFM-9485
	StreamRetry        SDKCode = "SDKCODE:5003"
	StreamStop         SDKCode = "SDKCODE:5004"
	EvaluationSuccess  SDKCode = "SDKCODE:6000"
	EvaluationFailed   SDKCode = "SDKCODE:6001"
	MissingBucketBy    SDKCode = "SDKCODE:6002"
	MetricsStarted     SDKCode = "SDKCODE:7000"
	MetricsStopped     SDKCode = "SDKCODE:7001"
	MetricsSendFail    SDKCode = "SDKCODE:7002"
	MetricsSendSuccess SDKCode = "SDKCODE:7003"
)

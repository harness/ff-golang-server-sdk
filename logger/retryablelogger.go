package logger

// RetryableLogger implements the Logger interface required by the go-retryablehttp client and wraps our internal logger
type RetryableLogger struct {
	logger Logger
}

// NewRetryableLogger creates a RetryableLogger instance.
func NewRetryableLogger(logger Logger) RetryableLogger {
	s := RetryableLogger{}
	s.logger = logger
	return s
}

// Printf - logs any printf messages from the http client as debug logs
func (s RetryableLogger) Printf(string string, args ...interface{}) {
	s.logger.Debugf(string, args...)
}

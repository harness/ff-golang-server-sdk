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

func (a RetryableLogger) Error(msg string, args ...interface{}) {
	a.logger.Error(msg, args)
}

func (a RetryableLogger) Info(msg string, args ...interface{}) {
	a.logger.Info(msg, args)
}

func (a RetryableLogger) Debug(msg string, args ...interface{}) {
	a.logger.Debug(msg, args)
}

func (a RetryableLogger) Warn(msg string, args ...interface{}) {
	a.logger.Warn(msg, args)
}

package logger

import "net/http"

// RetryableLogger implements the logger interface required by the go-retryablehttp client and wraps our internal logger
type RetryableLogger struct {
	logger Logger
}

// NewRetryableLogger creates a RetryableLogger instance.
func NewRetryableLogger(logger Logger) RetryableLogger {
	s := RetryableLogger{}
	s.logger = logger
	return s
}

//// Printf - logs any printf messages from the http client as debug logs
//func (s Retryablelogger) Printf(string string, args ...interface{}) {
//	s.logger.Debugf(string, args...)
//}

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

func LogRequestAttempt(logger Logger, req *http.Request, attemptNum int) {
	logger.Errorf("Attempt #%d: %s %s", attemptNum, req.Method, req.URL)
}

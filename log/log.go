package log

import "github.com/drone/ff-golang-server-sdk/logger"

// And just go global.
var defaultLogger logger.Logger

// init creates the default logger.  This can be changed
func init() {
	defaultLogger, _ = logger.NewZapLogger(false)
}

// SetLogger sets the default logger to be used by this package
func SetLogger(logger logger.Logger) {
	defaultLogger = logger
}

// Error logs an error message with the parameters
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatalf logs a formatted fatal message.  This will terminate the application
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Fatal logs an fatal message with the parameters.  This will terminate the application
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Info logs an info message with the parameters
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Warn logs an warn message with the parameters
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf logs a formatted warn message
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Debug logs an debug message with the parameters
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

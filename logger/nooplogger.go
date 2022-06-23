package logger

// NoOpLogger is a type that implements the Logger interface but does nothing
// when it's methods are called
type NoOpLogger struct{}

// NewNoOpLogger returns a NoOpLogger
func NewNoOpLogger() NoOpLogger {
	return NoOpLogger{}
}

// Debug does nothing on a NoOpLogger
func (m NoOpLogger) Debug(args ...interface{}) {}

// Debugf does nothing on a NoOpLogger
func (m NoOpLogger) Debugf(template string, args ...interface{}) {}

// Info does nothing on a NoOpLogger
func (m NoOpLogger) Info(args ...interface{}) {}

// Infof does nothing on a NoOpLogger
func (m NoOpLogger) Infof(template string, args ...interface{}) {}

// Warn does nothing on a NoOpLogger
func (m NoOpLogger) Warn(args ...interface{}) {}

// Warnf does nothing on a NoOpLogger
func (m NoOpLogger) Warnf(template string, args ...interface{}) {}

// Error does nothing on a NoOpLogger
func (m NoOpLogger) Error(args ...interface{}) {}

// Errorf does nothing on a NoOpLogger
func (m NoOpLogger) Errorf(template string, args ...interface{}) {}

// Panic does nothing on a NoOpLogger
func (m NoOpLogger) Panic(args ...interface{}) {}

// Panicf does nothing on a NoOpLogger
func (m NoOpLogger) Panicf(template string, args ...interface{}) {}

// Fatal does nothing on a NoOpLogger
func (m NoOpLogger) Fatal(args ...interface{}) {}

// Fatalf does nothing on a NoOpLogger
func (m NoOpLogger) Fatalf(template string, args ...interface{}) {}

package logger

type NoOpLogger struct{}

func NewNoOpLogger() NoOpLogger {
	return NoOpLogger{}
}

func (m NoOpLogger) Debug(args ...interface{}) {}

func (m NoOpLogger) Debugf(template string, args ...interface{}) {}

func (m NoOpLogger) Info(args ...interface{}) {}

func (m NoOpLogger) Infof(template string, args ...interface{}) {}

func (m NoOpLogger) Warn(args ...interface{}) {}

func (m NoOpLogger) Warnf(template string, args ...interface{}) {}

func (m NoOpLogger) Error(args ...interface{}) {}

func (m NoOpLogger) Errorf(template string, args ...interface{}) {}

func (m NoOpLogger) Panic(args ...interface{}) {}

func (m NoOpLogger) Panicf(template string, args ...interface{}) {}

func (m NoOpLogger) Fatal(args ...interface{}) {}

func (m NoOpLogger) Fatalf(template string, args ...interface{}) {}

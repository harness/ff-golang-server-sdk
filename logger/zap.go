package logger

import "go.uber.org/zap"

// ZapLogger object using simple SugaredLogger
type ZapLogger struct {
	logger *zap.SugaredLogger
}

// NewZapLogger creates a zap production or development logger based on debug argument
func NewZapLogger(debug bool) (*ZapLogger, error) {
	var (
		logger *zap.Logger
		err    error
	)

	if debug {
		logger, err = zap.NewDevelopment(zap.AddCallerSkip(1))
	} else {
		logger, err = zap.NewProduction(zap.AddCallerSkip(1))
	}
	if err != nil {
		return nil, err
	}
	sugar := logger.Sugar()
	return &ZapLogger{
		logger: sugar,
	}, nil
}

// NewZapLoggerFromSugar creates a ZapLogger from a zap.SugaredLogger
func NewZapLoggerFromSugar(sl zap.SugaredLogger) *ZapLogger {
	return &ZapLogger{logger: &sl}
}

// Debug uses zap to log a debug message.
func (z ZapLogger) Debug(args ...interface{}) {
	z.logger.Debug(args...)
}

// Debugf uses zap to log a debug message.
func (z ZapLogger) Debugf(template string, args ...interface{}) {
	z.logger.Debugf(template, args...)
}

// Info uses zap to log a info message.
func (z ZapLogger) Info(args ...interface{}) {
	z.logger.Info(args...)
}

// Infof uses zap to log a info message.
func (z ZapLogger) Infof(template string, args ...interface{}) {
	z.logger.Infof(template, args...)
}

// Warn uses zap to log a warning message.
func (z ZapLogger) Warn(args ...interface{}) {
	z.logger.Warn(args...)
}

// Warnf uses zap to log a warning message.
func (z ZapLogger) Warnf(template string, args ...interface{}) {
	z.logger.Warnf(template, args...)
}

// Error uses zap to log a error message.
func (z ZapLogger) Error(args ...interface{}) {
	z.logger.Error(args...)
}

// Panic uses zap to log a panic message.
func (z ZapLogger) Panic(args ...interface{}) {
	z.logger.Panic(args...)
}

// Panicf uses zap to log a panic message.
func (z ZapLogger) Panicf(template string, args ...interface{}) {
	z.logger.Panicf(template, args...)
}

// Fatal uses zap to log a fatal message.
func (z ZapLogger) Fatal(args ...interface{}) {
	z.logger.Fatal(args...)
}

// Fatalf uses zap to log a fatal message.
func (z ZapLogger) Fatalf(template string, args ...interface{}) {
	z.logger.Fatalf(template, args...)
}

// Errorf uses zap to log a error message.
func (z ZapLogger) Errorf(template string, args ...interface{}) {
	z.logger.Errorf(template, args...)
}

// Sugar returns the underlying sugared zap logger that ZapLogger uses
func (z ZapLogger) Sugar() *zap.SugaredLogger {
	return z.logger
}

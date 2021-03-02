package logger

import "go.uber.org/zap"

type ZapLogger struct {
	logger *zap.SugaredLogger
}

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

func (z ZapLogger) Debug(args ...interface{}) {
	z.logger.Debug(args...)
}

func (z ZapLogger) Debugf(template string, args ...interface{}) {
	z.logger.Debugf(template, args...)
}

func (z ZapLogger) Info(args ...interface{}) {
	z.logger.Info(args...)
}

func (z ZapLogger) Infof(template string, args ...interface{}) {
	z.logger.Infof(template, args...)
}

func (z ZapLogger) Warn(args ...interface{}) {
	z.logger.Warn(args...)
}

func (z ZapLogger) Warnf(template string, args ...interface{}) {
	z.logger.Warnf(template, args...)
}

func (z ZapLogger) Error(args ...interface{}) {
	z.logger.Error(args...)
}

func (z ZapLogger) Panic(args ...interface{}) {
	z.logger.Panic(args...)
}

func (z ZapLogger) Panicf(template string, args ...interface{}) {
	z.logger.Panicf(template, args...)
}

func (z ZapLogger) Fatal(args ...interface{}) {
	z.logger.Fatal(args...)
}

func (z ZapLogger) Fatalf(template string, args ...interface{}) {
	z.logger.Fatalf(template, args...)
}

func (z ZapLogger) Errorf(template string, args ...interface{}) {
	z.logger.Errorf(template, args...)
}

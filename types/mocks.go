package types

import (
	"fmt"
	"time"
)

type MockSleeper struct {
	SleepTime time.Duration
}

func (ms MockSleeper) Sleep(d time.Duration) {
	time.Sleep(ms.SleepTime)
}

type MockLogger struct {
	Logs []string
}

func (t MockLogger) Debug(args ...interface{}) { t.Logs = append(t.Logs, fmt.Sprint(args...)) }
func (t MockLogger) Debugf(template string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(template, args...))
}
func (t MockLogger) Info(args ...interface{}) { t.Logs = append(t.Logs, fmt.Sprint(args...)) }
func (t MockLogger) Infof(template string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(template, args...))
}
func (t MockLogger) Warn(args ...interface{}) { t.Logs = append(t.Logs, fmt.Sprint(args...)) }
func (t MockLogger) Warnf(template string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(template, args...))
}
func (t MockLogger) Error(args ...interface{}) { t.Logs = append(t.Logs, fmt.Sprint(args...)) }
func (t MockLogger) Errorf(template string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(template, args...))
}
func (t MockLogger) Panic(args ...interface{}) { t.Logs = append(t.Logs, fmt.Sprint(args...)) }
func (t MockLogger) Panicf(template string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(template, args...))
}
func (t MockLogger) Fatal(args ...interface{}) { t.Logs = append(t.Logs, fmt.Sprint(args...)) }
func (t MockLogger) Fatalf(template string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(template, args...))
}

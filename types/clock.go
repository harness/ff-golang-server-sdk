package types

import "time"

type Sleeper interface {
	Sleep(time.Duration)
}

// RealClock is the default implementation for Sleeper
type RealClock struct{}

func (rc RealClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

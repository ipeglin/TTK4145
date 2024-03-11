package timer

import (
	"time"
)

func GetCurrentTimeAsFloat() float64 {
	now := time.Now()
	// UnixNano returns nanoseconds since the Unix epoch.
	// To convert nanoseconds to seconds, divide by 1e9 (the number of nanoseconds in one second).
	return float64(now.Unix()) + float64(now.Nanosecond())*1e-9
}

var endTime float64
var isActive bool
var isInfinite bool

func Start(duration float64) {
	endTime = GetCurrentTimeAsFloat() + duration
	isActive = true
}

func StartInfiniteTimer() {
	isInfinite = true
	isActive = true
}

func StopInfiniteTimer() {
	isInfinite = false
	isActive = false
}

func Stop() {
	isActive = false
}

func TimerTimedOut() bool {
	return (!isInfinite && isActive && GetCurrentTimeAsFloat() > endTime)
}

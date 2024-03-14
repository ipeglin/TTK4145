package timer

import (
	"time"
)

func GetCurrentTimeAsFloat() float64 {
	now := time.Now()
	return float64(now.Unix()) + float64(now.Nanosecond())*1e-9
}

var endTime float64
var isActive bool
var IsInfinite bool

func TimedOut() bool {
	return (!IsInfinite && isActive && GetCurrentTimeAsFloat() > endTime)
}

func Start(duration float64) {
	endTime = GetCurrentTimeAsFloat() + duration
	isActive = true
}

func Stop() {
	isActive = false
}

func StartInfiniteTimer() {
	IsInfinite = true
	isActive = true
}

func StopInfiniteTimer() {
	IsInfinite = false
	isActive = false
}

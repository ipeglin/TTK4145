package timer

import (
	"time"
)

func GetWallTime() float64 {
	now := time.Now()
	// UnixNano returns nanoseconds since the Unix epoch.
	// To convert nanoseconds to seconds, divide by 1e9 (the number of nanoseconds in one second).
	return float64(now.Unix()) + float64(now.Nanosecond())*1e-9
}

var timerEndTime float64
var timerActive bool
var TimerInf bool

func TimerStart(duration float64) {
	timerEndTime = GetWallTime() + duration
	timerActive = true
}

func TimerStartInf() {
	TimerInf = true
	timerActive = true
}

func TimerStopInf() {
	TimerInf = false
	timerActive = false
}

func TimerStop() {
	timerActive = false
}

func TimerTimedOut() bool {
	return (!TimerInf && timerActive && GetWallTime() > timerEndTime)
}

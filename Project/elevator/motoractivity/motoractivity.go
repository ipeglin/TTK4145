package motoractivity

import (
	"elevator/elev"
	"elevator/elevio"
	"elevator/fsm"
	"elevator/timer"
	"time"
)

const motorTimeoutS = 3.0

func MontitorMotorActivity(receiver chan<- bool) {
	timerActive := true
	timerEndTimer := timer.GetCurrentTimeAsFloat() + motorTimeoutS
	v := elevio.RequestFloor()
	for {
		time.Sleep(elevio.PollRateMS * time.Millisecond)
		elevator := fsm.GetElevator()
		if v != -1 && (elevator.CurrentBehaviour != elev.EBMoving) || v != elevio.RequestFloor() {
			timerEndTimer = timer.GetCurrentTimeAsFloat() + motorTimeoutS
			if !timerActive {
				timerActive = true
				receiver <- true
			}
		} else {
			if timer.GetCurrentTimeAsFloat() > timerEndTimer {
				if timerActive {
					timerActive = false
					receiver <- false
				}
			}
		}
		v = elevio.RequestFloor()
	}
}

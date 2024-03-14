package motoractivity

import (
	"elevator/elevio"
	"elevator/fsm"
	"elevator/timer"
	"time"
)

const motorTimeoutS = 3.0

func MontitorMotorActivity(receiver chan<- bool) {
	timerActive := true
	timerEndTimer := timer.GetCurrentTimeAsFloat() + motorTimeoutS
	elevator := fsm.GetElevator()
	for {
		time.Sleep(elevio.PollRateMS * time.Millisecond)
		v := elevio.RequestFloor()
		if v != -1 && elevator.Dirn == elevio.DirStop {
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
	}
}

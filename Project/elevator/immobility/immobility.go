package immobility

import (
	"elevator/elev"
	"elevator/elevatorcontroller"
	"elevator/elevio"
	"elevator/statehandler"
	"elevator/timer"
	"time"
)

func MontitorMotorActivity(receiver chan<- bool) {
	timerActive := true
	timerEndTimer := timer.GetCurrentTimeAsFloat() + elev.MotorTimeoutS
	v := elevio.RequestFloor()
	for {
		time.Sleep(elevio.PollRateMS * time.Millisecond)
		elevator := elevatorcontroller.GetElevator()
		if v != -1 && (elevator.CurrentBehaviour != elev.EBMoving) || v != elevio.InputDevice.FloorSensor() {
			timerEndTimer = timer.GetCurrentTimeAsFloat() + elev.MotorTimeoutS
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
		v = elevio.InputDevice.FloorSensor()
	}
}

func RequestStartObstruction(elevatorName string) {
	elevator := elevatorcontroller.GetElevator()
	if elevator.CurrentBehaviour == elev.EBDoorOpen {
		timer.StartInfiniteTimer()
		statehandler.RemoveElevatorsFromState([]string{elevatorName})
	}
}

func StopObstruction() {
	elevator := elevatorcontroller.GetElevator()
	timer.StopInfiniteTimer()
	timer.Start(elevator.Config.DoorOpenDurationS)
}

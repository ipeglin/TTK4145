package counter

import (
	"elevator/elevio"
	"elevator/hra"
)

type Counter struct {
	HallRequests [][2]int       `json:"hallRequests"`
	States       map[string]int `json:"states"`
}

func InitialiseCounter(elevatorName string) Counter {
	counter := Counter{
		HallRequests: make([][2]int, elevio.NFloors),
		States:       make(map[string]int),
	}

	counter.States[elevatorName] = 0

	return counter
}

func UpdateOnCompletedOrder(counter Counter, elevatorName string, btn_floor int, btn_type elevio.Button) Counter {
	switch btn_type {
	case elevio.BHallUp:
		counter.HallRequests[btn_floor][elevio.BHallUp] += 1
	case elevio.BHallDown:
		counter.HallRequests[btn_floor][elevio.BHallDown] += 1
	}
	counter.States[elevatorName] += 1
	return counter
}

func IncrementOnInput(counter Counter, elevatorName string) Counter {
	counter.States[elevatorName] += 1
	return counter
}

func UpdateOnNewOrder(counter Counter, hraInput hra.HRAInput, elevatorName string, btnFloor int, btn elevio.Button) Counter {
	switch btn {
	case elevio.BHallUp:
		if !hraInput.HallRequests[btnFloor][elevio.BHallUp] {
			counter.HallRequests[btnFloor][elevio.BHallUp] += 1
		}
	case elevio.BHallDown:
		if !hraInput.HallRequests[btnFloor][elevio.BHallDown] {
			counter.HallRequests[btnFloor][elevio.BHallDown] += 1
		}
	case elevio.BCab:
		counter.States[elevatorName] += 1
	}
	return counter
}

package counter

import (
	"elevator/elevio"
	"elevator/hra"
)

// TODO: Change to just CyclicCounter
type Counter struct {
	HallRequests [][2]int       `json:"hallRequests"`
	States       map[string]int `json:"states"`
}

func InitializeCounter(elevatorName string) Counter {
	cyclicCounter := Counter{
		HallRequests: make([][2]int, elevio.NFloors),
		States:       make(map[string]int), // Initialiserer map her
	}

	// Nå som States er initialisert, kan du legge til oppføringer i den
	cyclicCounter.States[elevatorName] = 0

	return cyclicCounter
}

func UpdateOnCompletedOrder(cyclicCounter Counter, elevatorName string, btn_floor int, btn_type elevio.Button) Counter {
	switch btn_type {
	case elevio.BHallUp:
		cyclicCounter.HallRequests[btn_floor][0] += 1
	case elevio.BHallDown:
		cyclicCounter.HallRequests[btn_floor][1] += 1
	}
	cyclicCounter.States[elevatorName] += 1
	return cyclicCounter
}

func IncrementOnInput(cyclicCounter Counter, elevatorName string) Counter {
	cyclicCounter.States[elevatorName] += 1
	return cyclicCounter
}

func UpdateOnNewOrder(cyclicCounter Counter, hraInput hra.HRAInput, elevatorName string, btnFloor int, btn elevio.Button) Counter {
	switch btn {
	case elevio.BHallUp:
		if !hraInput.HallRequests[btnFloor][0] {
			cyclicCounter.HallRequests[btnFloor][0] += 1
		}
	case elevio.BHallDown:
		if !hraInput.HallRequests[btnFloor][1] {
			cyclicCounter.HallRequests[btnFloor][1] += 1
		}
	case elevio.BCab:
		cyclicCounter.States[elevatorName] += 1
	}
	return cyclicCounter
}

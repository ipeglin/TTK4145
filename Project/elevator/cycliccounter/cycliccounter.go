package cycliccounter

import (
	"elevator/elevio"
	"elevator/hra"
)

// TODO: Change to just CyclicCounter
type CyclicCounterInput struct {
	HallRequests [][2]int       `json:"hallRequests"`
	States       map[string]int `json:"states"`
}

func InitializeCyclicCounterInput(ElevatorName string) CyclicCounterInput {
	cyclicCounter := CyclicCounterInput{
		HallRequests: make([][2]int, elevio.NFloors),
		States:       make(map[string]int), // Initialiserer map her
	}

	// Nå som States er initialisert, kan du legge til oppføringer i den
	cyclicCounter.States[ElevatorName] = 0

	return cyclicCounter
}

func UpdateOnCompletedOrder(cyclicCounter CyclicCounterInput, elevatorName string, btn_floor int, btn_type elevio.Button) CyclicCounterInput {
	switch btn_type {
	case elevio.BHallUp:
		cyclicCounter.HallRequests[btn_floor][0] += 1
	case elevio.BHallDown:
		cyclicCounter.HallRequests[btn_floor][1] += 1
	}
	cyclicCounter.States[elevatorName] += 1
	return cyclicCounter
}

func IncrementOnInput(cyclicCounter CyclicCounterInput, elevatorName string) CyclicCounterInput {
	cyclicCounter.States[elevatorName] += 1
	return cyclicCounter
}

func UpdateOnNewOrder(cyclicCounter CyclicCounterInput, hraInput hra.HRAInput, elevatorName string, btnFloor int, btn elevio.Button) CyclicCounterInput {
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

func synchronizeLocalCCounterWithIncoming(localCCounter *CyclicCounterInput, IncomingCCounter CyclicCounterInput, localElevatorName string) {
	if _, exists := IncomingCCounter.States[localElevatorName]; exists {
		if IncomingCCounter.States[localElevatorName] > localCCounter.States[localElevatorName] {
			localCCounter.States[localElevatorName] = IncomingCCounter.States[localElevatorName] + 1
		}
	}
}
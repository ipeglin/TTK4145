package hra

import (
	"elevator/elev"
	"elevator/elevio"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func InitializeHRAInput(e elev.Elevator, elevatorName string) HRAInput {
	hraInput := HRAInput{
		HallRequests: make([][2]bool, elevio.NFloors),
		States:       make(map[string]HRAElevState),
	}
	for f := 0; f < elevio.NFloors; f++ {
		hraInput.HallRequests[f][0] = e.Requests[f][elevio.BHallUp]
		hraInput.HallRequests[f][1] = e.Requests[f][elevio.BHallDown]
	}

	behavior, direction, cabRequests := convertElevatorState(e)

	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       e.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

func UpdateHRAInput(hraInput HRAInput, e elev.Elevator, elevatorName string) HRAInput {
	for f := 0; f < elevio.NFloors; f++ {
		hraInput.HallRequests[f][0] = hraInput.HallRequests[f][0] || e.Requests[f][elevio.BHallUp]
		hraInput.HallRequests[f][1] = hraInput.HallRequests[f][1] || e.Requests[f][elevio.BHallDown]
	}

	behavior, direction, cabRequests := convertElevatorState(e)
	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       e.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

func RebootHRAInput(hraInput HRAInput, e elev.Elevator, elevatorName string) HRAInput {
	behavior, direction, cabRequests := convertElevatorState(e)
	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       e.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

func UpdateHRAInputOnCompletedOrder(hraInput HRAInput, e elev.Elevator, elevatorName string, btn_floor int, btn_type elevio.Button) HRAInput {
	switch btn_type {
	case elevio.BHallUp:
		hraInput.HallRequests[btn_floor][0] = false
	case elevio.BHallDown:
		hraInput.HallRequests[btn_floor][1] = false
	case elevio.BCab:
		hraInput.States[elevatorName].CabRequests[btn_floor] = false
	}

	behavior, direction, cabRequests := convertElevatorState(e)

	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       e.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

func convertElevatorState(e elev.Elevator) (string, string, []bool) {
	behavior := elev.EBToString(e.CurrentBehaviour)
	direction := elevio.ElevDirToString(e.Dirn)

	// Convert cab requests
	cabRequests := make([]bool, elevio.NFloors)
	for f := 0; f < elevio.NFloors; f++ {
		cabRequests[f] = e.Requests[f][elevio.BCab]
	}

	return behavior, direction, cabRequests
}

func UpdateHRAInputOnNewOrder(hraInput HRAInput, elevatorName string, btnFloor int, btn elevio.Button) HRAInput {
	switch btn {
	case elevio.BHallUp:
		hraInput.HallRequests[btnFloor][0] = true
	case elevio.BHallDown:
		hraInput.HallRequests[btnFloor][1] = true
	case elevio.BCab:
		hraInput.States[elevatorName].CabRequests[btnFloor] = true
	}
	return hraInput
}

package hra

import (
	"elevator/elev"
	"elevator/elevio"
)

const FilenameHRAInput = "elevHRAInput.json"

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

func InitializeHRAInput(el elev.Elevator, elevatorName string) HRAInput {
	hraInput := HRAInput{
		HallRequests: make([][2]bool, elevio.NFloors),
		States:       make(map[string]HRAElevState),
	}
	for f := 0; f < elevio.NFloors; f++ {
		hraInput.HallRequests[f][0] = el.Requests[f][elevio.BHallUp]
		hraInput.HallRequests[f][1] = el.Requests[f][elevio.BHallDown]
	}

	behavior, direction, cabRequests := convertLocalElevatorState(el)

	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       el.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

func UpdateHRAInput(hraInput HRAInput, el elev.Elevator, elevatorName string) HRAInput {
	for f := 0; f < elevio.NFloors; f++ {
		hraInput.HallRequests[f][0] = hraInput.HallRequests[f][0] || el.Requests[f][elevio.BHallUp]
		hraInput.HallRequests[f][1] = hraInput.HallRequests[f][1] || el.Requests[f][elevio.BHallDown]
	}
	
	behavior, direction, cabRequests := convertLocalElevatorState(el)
	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       el.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
	}

func RebootHRAInput(hraInput HRAInput, el elev.Elevator, elevatorName string) HRAInput {
	behavior, direction, cabRequests := convertLocalElevatorState(el)
	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       el.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}


func UpdateHRAInputOnCompletedOrder(hraInput HRAInput, el elev.Elevator, elevatorName string, btn_floor int, btn_type elevio.Button) HRAInput {
	switch btn_type {
	case elevio.BHallUp:
		hraInput.HallRequests[btn_floor][0] = false
	case elevio.BHallDown:
		hraInput.HallRequests[btn_floor][1] = false
	case elevio.BCab:
		hraInput.States[elevatorName].CabRequests[btn_floor] = false
	}

	behavior, direction, cabRequests := convertLocalElevatorState(el)

	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       el.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

//HAR SIMEN FUNC FOR DETTE ALLERDE? 
func convertLocalElevatorState(localElevator elev.Elevator) (string, string, []bool) {
	// Convert behavior
	var behavior string
	switch localElevator.CurrentBehaviour {
	case elev.EBIdle:
		behavior = "idle"
	case elev.EBMoving:
		behavior = "moving"
	case elev.EBDoorOpen:
		behavior = "doorOpen"
	}
	// Convert direction
	var direction string
	switch localElevator.Dirn {
	case elevio.DirUp:
		direction = "up"
	case elevio.DirDown:
		direction = "down"
	default:
		direction = "stop"
	}

	// Convert cab requests
	cabRequests := make([]bool, elevio.NFloors)
	for f := 0; f < elevio.NFloors; f++ {
		cabRequests[f] = localElevator.Requests[f][elevio.BCab]
	}

	return behavior, direction, cabRequests
}

func UpdateHRAInputWhenNewOrderOccurs(hraInput HRAInput, elevatorName string, btnFloor int, btn elevio.Button) HRAInput {
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


func synchronizeLocalHRAWithIncoming(localCombinedInput *jsonhandler.CombinedInput, otherCombinedInput jsonhandler.CombinedInput, incomingElevatorName string, localElevatorName string) {
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] > localCombinedInput.CyclicCounter.HallRequests[f][i] {
				localCombinedInput.CyclicCounter.HallRequests[f][i] = otherCombinedInput.CyclicCounter.HallRequests[f][i]
				localCombinedInput.HRAInput.HallRequests[f][i] = otherCombinedInput.HRAInput.HallRequests[f][i]
			}
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] == localCombinedInput.CyclicCounter.HallRequests[f][i] {
				if localCombinedInput.HRAInput.HallRequests[f][i] != otherCombinedInput.HRAInput.HallRequests[f][i] {
					localCombinedInput.HRAInput.HallRequests[f][i] = false
				}
			}
		}
	}

	if _, exists := otherCombinedInput.HRAInput.States[incomingElevatorName]; exists {
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; !exists {
			localCombinedInput.HRAInput.States[incomingElevatorName] = otherCombinedInput.HRAInput.States[incomingElevatorName]
			localCombinedInput.CyclicCounter.States[incomingElevatorName] = otherCombinedInput.CyclicCounter.States[incomingElevatorName]
		} else {
			if otherCombinedInput.CyclicCounter.States[incomingElevatorName] > localCombinedInput.CyclicCounter.States[incomingElevatorName] {
				localCombinedInput.HRAInput.States[incomingElevatorName] = otherCombinedInput.HRAInput.States[incomingElevatorName]
				localCombinedInput.CyclicCounter.States[incomingElevatorName] = otherCombinedInput.CyclicCounter.States[incomingElevatorName]
			}
		}
	} else {
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; exists {
			delete(localCombinedInput.HRAInput.States, incomingElevatorName)
			delete(localCombinedInput.CyclicCounter.States, incomingElevatorName)
		}
	}
}
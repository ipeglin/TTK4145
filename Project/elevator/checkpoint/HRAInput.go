package checkpoint

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

func initializeHRAInput(el elev.Elevator, elevatorName string) HRAInput {
	// Create a default HRAInput. Modify this according to your requirements.
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

func updateHRAInput(hraInput HRAInput, el elev.Elevator, elevatorName string) HRAInput {
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


func updateHRAInputWhenOrderIsComplete(hraInput HRAInput, el elev.Elevator, elevatorName string, btn_floor int, btn_type elevio.Button) HRAInput {
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

func updateHRAInputWhenNewOrderOccurs(hraInput HRAInput, elevatorName string, btnFloor int, btn elevio.Button, localElevator *elev.Elevator) HRAInput {
	switch btn {
	case elevio.BHallUp:
		hraInput.HallRequests[btnFloor][0] = true
	case elevio.BHallDown:
		hraInput.HallRequests[btnFloor][1] = true
	case elevio.BCab:
		hraInput.States[elevatorName].CabRequests[btnFloor] = true
		//denne burde ikke endres inne i her, men krever ellers så mange if setninger i kode. tenk smart løsning
		//localElevator.Requests[btnFloor][btn] = true
	}
	return hraInput
}

package requests

import (
	//"fmt"
	"elevator/checkpoint"
	"elevator/elev"
	"elevator/elevio"
)

func requestsAbove(e elev.Elevator) bool {
	for f := e.CurrentFloor + 1; f < elevio.NFloors; f++ {
		for btn := 0; btn < elevio.NButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elev.Elevator) bool {
	for f := 0; f < e.CurrentFloor; f++ {
		for btn := 0; btn < elevio.NButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elev.Elevator) bool {
	for btn := 0; btn < elevio.NButtons; btn++ {
		if e.Requests[e.CurrentFloor][btn] {
			return true

		}
	}
	return false
}

func decideDirectionUp(e elev.Elevator) DirnBehaviourPair {
	if requestsAbove(e) {
		return DirnBehaviourPair{elevio.DirUp, elev.EBMoving}
	} else if requestsHere(e) {
		return DirnBehaviourPair{elevio.DirDown, elev.EBDoorOpen}
	} else if requestsBelow(e) {
		return DirnBehaviourPair{elevio.DirDown, elev.EBMoving}
	}
	return DirnBehaviourPair{elevio.DirStop, elev.EBIdle}
}

func decideDirectionDown(e elev.Elevator) DirnBehaviourPair {
	if requestsBelow(e) {
		return DirnBehaviourPair{elevio.DirDown, elev.EBMoving}
	} else if requestsHere(e) {
		return DirnBehaviourPair{elevio.DirUp, elev.EBDoorOpen}
	} else if requestsAbove(e) {
		return DirnBehaviourPair{elevio.DirUp, elev.EBMoving}
	}
	return DirnBehaviourPair{elevio.DirStop, elev.EBIdle}
}

func decideDirectionStop(e elev.Elevator) DirnBehaviourPair {
	if requestsHere(e) {
		return DirnBehaviourPair{elevio.DirStop, elev.EBDoorOpen}
	} else if requestsAbove(e) {
		return DirnBehaviourPair{elevio.DirUp, elev.EBMoving}
	} else if requestsBelow(e) {
		return DirnBehaviourPair{elevio.DirDown, elev.EBMoving}
	}
	return DirnBehaviourPair{elevio.DirStop, elev.EBIdle}
}

func ChooseDirection(e elev.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevio.DirUp:
		return decideDirectionUp(e)
	case elevio.DirDown:
		return decideDirectionDown(e)
	case elevio.DirStop:
		return decideDirectionStop(e)
	default:
		return DirnBehaviourPair{
			elevio.DirStop,
			elev.EBIdle,
		}
	}
}

func ShouldStop(e elev.Elevator) bool {
	switch e.Dirn {
	case elevio.DirDown:
		return e.Requests[e.CurrentFloor][elevio.BHallDown] ||
			e.Requests[e.CurrentFloor][elevio.BCab] ||
			!requestsBelow(e)
	case elevio.DirUp:
		return e.Requests[e.CurrentFloor][elevio.BHallUp] ||
			e.Requests[e.CurrentFloor][elevio.BCab] ||
			!requestsAbove(e)
	default:
		return true
	}
}

func ShouldClearImmediately(e elev.Elevator, btn_floor int, btn_type elevio.Button) bool {
	switch e.Config.ClearRequestVariant {
	case elev.CRVAll:
		return e.CurrentFloor == btn_floor

	case elev.CRVInDirn:
		return e.CurrentFloor == btn_floor &&
			((e.Dirn == elevio.DirUp && btn_type == elevio.BHallUp) ||
				(e.Dirn == elevio.DirDown && btn_type == elevio.BHallDown) ||
				e.Dirn == elevio.DirStop ||
				btn_type == elevio.BCab)
	default:
		return false
	}
}

func ClearAtCurrentFloor(e elev.Elevator, filename string, elevatorName string) elev.Elevator {

	beforeClear := make(map[elevio.Button]bool)
	for btn := 0; btn < elevio.NButtons; btn++ {
		beforeClear[elevio.Button(btn)] = e.Requests[e.CurrentFloor][btn]
	}
	switch e.Config.ClearRequestVariant {
	case elev.CRVAll:
		for btn := 0; btn < elevio.NButtons; btn++ {
			e.Requests[e.CurrentFloor][btn] = false
		}

	case elev.CRVInDirn:
		e.Requests[e.CurrentFloor][elevio.BCab] = false
		switch e.Dirn {
		case elevio.DirUp:
			if !requestsAbove(e) && !e.Requests[e.CurrentFloor][elevio.BHallUp] {
				e.Requests[e.CurrentFloor][elevio.BHallDown] = false
			}
			e.Requests[e.CurrentFloor][elevio.BHallUp] = false

		case elevio.DirDown:
			if !requestsBelow(e) && !e.Requests[e.CurrentFloor][elevio.BHallDown] {
				e.Requests[e.CurrentFloor][elevio.BHallUp] = false
			}
			e.Requests[e.CurrentFloor][elevio.BHallDown] = false
		default:
			e.Requests[e.CurrentFloor][elevio.BHallUp] = false
			e.Requests[e.CurrentFloor][elevio.BHallDown] = false

		}
	}
	for btn, wasPressed := range beforeClear {
		if wasPressed && !e.Requests[e.CurrentFloor][btn] {
			checkpoint.UpdateJSONOnCompletedHallOrder(e, filename, elevatorName, e.CurrentFloor, btn)
		}
	}
	return e
}

func ClearAll(e *elev.Elevator) {
	for f := 0; f < elevio.NFloors; f++ {
		for btn := 0; btn < elevio.NButtons; btn++ {
			e.Requests[f][btn] = false
		}
	}
}

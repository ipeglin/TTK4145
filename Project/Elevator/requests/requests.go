package requests

import (
	"heislab/Elevator/elevator"
	"heislab/Elevator/elevio"
)

func requestsAbove(e elevator.Elevator) bool {
	for f := e.CurrentFloor + 1; f < elevio.NFloors; f++ {
		for btn := 0; btn < elevio.NButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevator.Elevator) bool {
	for f := 0; f < e.CurrentFloor; f++ {
		for btn := 0; btn < elevio.NButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elevator.Elevator) bool {
	for btn := 0; btn < elevio.NButtons; btn++ {
		if e.Requests[e.CurrentFloor][btn] {
			return true

		}
	}
	return false
}

func decideDirectionUp(e elevator.Elevator) DirnBehaviourPair {
	if requestsAbove(e) {
		return DirnBehaviourPair{elevio.DirUp, elevator.EBMoving}
	} else if requestsHere(e) {
		return DirnBehaviourPair{elevio.DirDown, elevator.EBDoorOpen}
	} else if requestsBelow(e) {
		return DirnBehaviourPair{elevio.DirDown, elevator.EBMoving}
	}
	return DirnBehaviourPair{elevio.DirStop, elevator.EBIdle}
}

func decideDirectionDown(e elevator.Elevator) DirnBehaviourPair {
	if requestsBelow(e) {
		return DirnBehaviourPair{elevio.DirDown, elevator.EBMoving}
	} else if requestsHere(e) {
		return DirnBehaviourPair{elevio.DirUp, elevator.EBDoorOpen}
	} else if requestsAbove(e) {
		return DirnBehaviourPair{elevio.DirUp, elevator.EBMoving}
	}
	return DirnBehaviourPair{elevio.DirStop, elevator.EBIdle}
}

func decideDirectionStop(e elevator.Elevator) DirnBehaviourPair {
	if requestsHere(e) {
		return DirnBehaviourPair{elevio.DirStop, elevator.EBDoorOpen}
	} else if requestsAbove(e) {
		return DirnBehaviourPair{elevio.DirUp, elevator.EBMoving}
	} else if requestsBelow(e) {
		return DirnBehaviourPair{elevio.DirDown, elevator.EBMoving}
	}
	return DirnBehaviourPair{elevio.DirStop, elevator.EBIdle}
}

func requestsChooseDirection(e elevator.Elevator) DirnBehaviourPair {
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
			elevator.EBIdle,
		}
	}
}

func requestsShouldStop(e elevator.Elevator) bool {
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

func requestsShouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.Button) bool {
	switch e.Config.ClearRequestVariant {
	case elevator.CRVAll:
		return e.CurrentFloor == btn_floor
	case elevator.CRVInDirn:
		return e.CurrentFloor == btn_floor &&
			((e.Dirn == elevio.DirUp && btn_type == elevio.BHallUp) ||
				(e.Dirn == elevio.DirDown && btn_type == elevio.BHallDown) ||
				e.Dirn == elevio.DirStop ||
				btn_type == elevio.BCab)
	default:
		return false
	}
}

func requestsClearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {
	switch e.Config.ClearRequestVariant {
	case elevator.CRVAll:
		for btn := 0; btn < elevio.NButtons; btn++ {
			e.Requests[e.CurrentFloor][btn] = false
		}

	case elevator.CRVInDirn:
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
			e.Requests[e.CurrentFloor][elevio.BHallUp] = false

		default:
			e.Requests[e.CurrentFloor][elevio.BHallUp] = false
			e.Requests[e.CurrentFloor][elevio.BHallDown] = false
		}
	default:

	}
	return e
}

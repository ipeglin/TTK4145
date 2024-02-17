package requests

import (
	"heislab/Elevator/driver/hwelevio"
	"heislab/Elevator/elev"
	"heislab/Elevator/elevio"
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

func RequestsChooseDirection(e elev.Elevator) DirnBehaviourPair {
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

func RequestsShouldStop(e elev.Elevator) bool {
	switch e.Dirn {
	case elevio.DirDown:
		return e.Requests[e.CurrentFloor][hwelevio.BHallDown] ||
			e.Requests[e.CurrentFloor][hwelevio.BCab] ||
			!requestsBelow(e)
	case elevio.DirUp:
		return e.Requests[e.CurrentFloor][hwelevio.BHallUp] ||
			e.Requests[e.CurrentFloor][hwelevio.BCab] ||
			!requestsAbove(e)
	default:
		return true
	}
}

func RequestsShouldClearImmediately(e elev.Elevator, btn_floor int, btn_type hwelevio.Button) bool {
	switch e.Config.ClearRequestVariant {
	case elev.CRVAll:
		return e.CurrentFloor == btn_floor
	case elev.CRVInDirn:
		return e.CurrentFloor == btn_floor &&
			((e.Dirn == elevio.DirUp && btn_type == hwelevio.BHallUp) ||
				(e.Dirn == elevio.DirDown && btn_type == hwelevio.BHallDown) ||
				e.Dirn == elevio.DirStop ||
				btn_type == hwelevio.BCab)
	default:
		return false
	}
}

func RequestsClearAtCurrentFloor(e elev.Elevator) elev.Elevator {
	switch e.Config.ClearRequestVariant {
	case elev.CRVAll:
		for btn := 0; btn < elevio.NButtons; btn++ {
			e.Requests[e.CurrentFloor][btn] = false
		}

	case elev.CRVInDirn:
		e.Requests[e.CurrentFloor][hwelevio.BCab] = false
		switch e.Dirn {
		case elevio.DirUp:
			if !requestsAbove(e) && !e.Requests[e.CurrentFloor][hwelevio.BHallUp] {
				e.Requests[e.CurrentFloor][hwelevio.BHallDown] = false
			}
			e.Requests[e.CurrentFloor][hwelevio.BHallUp] = false

		case elevio.DirDown:
			if !requestsBelow(e) && !e.Requests[e.CurrentFloor][hwelevio.BHallDown] {
				e.Requests[e.CurrentFloor][hwelevio.BHallUp] = false
			}
			e.Requests[e.CurrentFloor][hwelevio.BHallUp] = false

		default:
			e.Requests[e.CurrentFloor][hwelevio.BHallUp] = false
			e.Requests[e.CurrentFloor][hwelevio.BHallDown] = false
		}
	default:

	}
	return e
}

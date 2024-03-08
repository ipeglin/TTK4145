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

func RequestsShouldClearImmediately(e elev.Elevator, btn_floor int, btn_type elevio.Button) bool {
	switch e.Config.ClearRequestVariant {
	case elev.CRVAll:
		//fmt.Print("CRVAll, RequestsShouldClearImmediately")
		return e.CurrentFloor == btn_floor
	case elev.CRVInDirn:
		//fmt.Print("CRVInDirn, RequestsShouldClearImmediately")
		return e.CurrentFloor == btn_floor &&
			((e.Dirn == elevio.DirUp && btn_type == elevio.BHallUp) ||
				(e.Dirn == elevio.DirDown && btn_type == elevio.BHallDown) ||
				e.Dirn == elevio.DirStop ||
				btn_type == elevio.BCab)
	default:
		return false
	}
}

func RequestsClearAtCurrentFloor(e elev.Elevator, filename string, elevatorName string) elev.Elevator {
	//fmt.Print("RequestsClearAtCurrentFloor: ")
	switch e.Config.ClearRequestVariant {
	case elev.CRVAll:
		//fmt.Print("CRVAll, RequestsClearAtCurrentFloor")
		for btn := 0; btn < elevio.NButtons; btn++ {
			e.Requests[e.CurrentFloor][btn] = false
		}

	case elev.CRVInDirn:
		//fmt.Print("CRVInDirn, RequestsClearAtCurrentFloor")
		e.Requests[e.CurrentFloor][elevio.BCab] = false
		checkpoint.UpdateJSONWhenHallOrderIsComplete(e, filename, elevatorName, e.CurrentFloor, elevio.BCab)
		switch e.Dirn {
		case elevio.DirUp:
			if !requestsAbove(e) && !e.Requests[e.CurrentFloor][elevio.BHallUp] {
				e.Requests[e.CurrentFloor][elevio.BHallDown] = false
				checkpoint.UpdateJSONWhenHallOrderIsComplete(e, filename, elevatorName, e.CurrentFloor, elevio.BHallDown)
			}
			e.Requests[e.CurrentFloor][elevio.BHallUp] = false
			checkpoint.UpdateJSONWhenHallOrderIsComplete(e, filename, elevatorName, e.CurrentFloor,elevio.BHallUp)

		case elevio.DirDown:
			if !requestsBelow(e) && !e.Requests[e.CurrentFloor][elevio.BHallDown] {
				e.Requests[e.CurrentFloor][elevio.BHallUp] = false
				checkpoint.UpdateJSONWhenHallOrderIsComplete(e, filename, elevatorName, e.CurrentFloor,elevio.BHallUp)
			}
			e.Requests[e.CurrentFloor][elevio.BHallUp] = false
			checkpoint.UpdateJSONWhenHallOrderIsComplete(e, filename, elevatorName, e.CurrentFloor,elevio.BHallUp)

		default:
			e.Requests[e.CurrentFloor][elevio.BHallUp] = false
			checkpoint.UpdateJSONWhenHallOrderIsComplete(e, filename, elevatorName, e.CurrentFloor,elevio.BHallUp)
			e.Requests[e.CurrentFloor][elevio.BHallDown] = false
			checkpoint.UpdateJSONWhenHallOrderIsComplete(e, filename, elevatorName, e.CurrentFloor,elevio.BHallDown)
		}
	default:

	}
	return e
}

func RequestsClearAll(e *elev.Elevator) {
	for f := 0; f < elevio.NFloors; f++ {
		for btn := 0; btn < elevio.NButtons; btn++ {
			e.Requests[f][btn] = false
		}
	}
}

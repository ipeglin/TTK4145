package fsm

import (
	"heislab/Elevator/elev"
	"heislab/Elevator/elevio"
	"heislab/Elevator/requests"
	"heislab/Elevator/timer"
)

/* "heislab/Elevator/elevio"
"heislab/Elevator/timer"

"heislab/Elevator/requests" */

var elevator elev.Elevator
var outputDevice elevio.ElevOutputDevice

func init() {
	elevator = elev.ElevatorInit()
	//TODO
	outputDevice = elevio.ElevioGetOutputDevice()
}

func setAllLights(e elev.Elevator) {
	for floor := 0; floor < elevio.NFloors; floor++ {
		for btn := 0; btn < elevio.NButtons; btn++ {
			outputDevice.RequestButtonLight(floor, elevio.Button(btn), e.Requests[floor][btn])
		}
	}
}

func fsmInitBetweenFloors() {
	dirn := elevio.DirDown
	outputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

func fsmRequestButtonPress(btn_floor int, btn_type elevio.Button) {
	//BUrde være tekst her mulgiens
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		if requests.RequestsShouldClearImmediately(elevator, btn_floor, btn_type) {
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
		} else {
			elevator.Requests[btn_floor][btn_type] = true
		}

	case elev.EBMoving:
		elevator.Requests[btn_floor][btn_type] = true

	case elev.EBIdle:
		elevator.Requests[btn_floor][btn_type] = true
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)

		case elev.EBMoving:
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights(elevator)
	//Mer tekst?
}

func fsmFloorArrival(newFloor int) {
	//print?
	elevator.CurrentFloor = newFloor

	outputDevice.FloorIndicator(elevator.CurrentFloor)
	//Helt unødvendig med switch her?
	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.RequestsShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			outputDevice.DoorLight(true)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			setAllLights(elevator)
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}
	//Mer print?
}

func fsmDoorTimeout() {
	//print?
	//Hvorfor switch
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch elevator.CurrentBehaviour {
		case elev.EBDoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)
			setAllLights(elevator)

		case elev.EBIdle:
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	//Mer print
}

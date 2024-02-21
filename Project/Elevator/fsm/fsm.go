package fsm

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
	"heislab/Elevator/elev"
	"heislab/Elevator/elevio"
	"heislab/Elevator/requests"
	"heislab/Elevator/timer"
)

var elevator elev.Elevator
var outputDevice elevio.ElevOutputDevice

func init() {
	elevator = elev.ElevatorInit()
	fmt.Println("fsm_init has happend")
	//TODO
	outputDevice = elevio.ElevioGetOutputDevice()
	//Burde dette gå et annet sted?
	setAllLights(elevator)
	hwelevio.SetDoorOpenLamp(false)
	hwelevio.SetStopLamp(false)
}

func setAllLights(e elev.Elevator) {
	for floor := 0; floor < elevio.NFloors; floor++ {
		for btn := hwelevio.BHallUp; btn <= hwelevio.BCab; btn++ {
			outputDevice.RequestButtonLight(floor, btn, e.Requests[floor][btn])
		}
	}
}

func FsmInitBetweenFloors() {
	dirn := elevio.DirDown
	fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(dirn), " in FsmInitBetweenFloors")
	outputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

func FsmRequestButtonPress(btnFloor int, btnType hwelevio.Button) {

	fmt.Printf("\n\n%s(%d, %s)\n", "FsmRequestButtonPress", btnFloor, hwelevio.ButtonToString(btnType))
	elev.ElevatorPrint(elevator)

	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		if requests.RequestsShouldClearImmediately(elevator, btnFloor, btnType) {
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
		} else {
			elevator.Requests[btnFloor][btnType] = true
		}

	case elev.EBMoving:
		elevator.Requests[btnFloor][btnType] = true

	case elev.EBIdle:
		elevator.Requests[btnFloor][btnType] = true
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)

		case elev.EBMoving:
			fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevator.Dirn), " in FsmRequestButtonPress")
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights(elevator)
	fmt.Printf("New state: \n")
	elev.ElevatorPrint(elevator)
}

func FsmFloorArrival(newFloor int) {
	fmt.Printf("\n\n%s(%d)\n", "FsmFloorArrival", newFloor)
	elev.ElevatorPrint(elevator)
	elevator.CurrentFloor = newFloor
	outputDevice.FloorIndicator(elevator.CurrentFloor)
	//Helt unødvendig med switch her?
	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.RequestsShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			elevator.Dirn = elevio.DirStop
			outputDevice.DoorLight(true)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			setAllLights(elevator)
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}
	fmt.Println("New state:")
	elev.ElevatorPrint(elevator)
}

func FsmDoorTimeout() {
	fmt.Printf("\n\n%s()\n", "FsmDoorTimeout")
	elev.ElevatorPrint(elevator)
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

		case elev.EBMoving:
			outputDevice.DoorLight(false)
			fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevio.DirStop), " in FsmDoorTimeout")
			outputDevice.MotorDirection(elevator.Dirn)
		case elev.EBIdle:
			outputDevice.DoorLight(false)
		}

	}
	fmt.Println("New State: \n")
	elev.ElevatorPrint(elevator)
}

// TODO
func FsmObstruction() {
	if elevator.CurrentBehaviour == elev.EBDoorOpen {
		timer.TimerStart(elevator.Config.DoorOpenDurationS)
	}

}

// TODO
func FsmStop(stop bool) {
	hwelevio.SetStopLamp(stop)
}

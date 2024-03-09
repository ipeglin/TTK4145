package fsm

import (
	"elevator/checkpoint"
	"elevator/elev"
	"elevator/elevio"
	"elevator/requests"
	"elevator/timer"
	"fmt"
	"network/local"
	"os"
	"time"
)

var elevator elev.Elevator
var outputDevice elevio.ElevOutputDevice
var localIP string

func init() {
	elevator = elev.ElevatorInit()
	localIP, _ = local.GetIP()
	//fmt.Println("fsm_init has happend")
	//TODO
	outputDevice = elevio.ElevioGetOutputDevice()
	//Burde dette gå et annet sted?
	setAllLights()
	elevio.RequestDoorOpenLamp(false)
	elevio.RequestStopLamp(false)
}

// init og denne vil kræsje ved process pair må finnes ut av
func SetElevator(f int, cb elev.ElevatorBehaviour, dirn elevio.ElevDir, r [elevio.NFloors][elevio.NButtons]bool, c elev.ElevatorConfig) {
	elevator.CurrentFloor = f
	elevator.CurrentBehaviour = cb
	elevator.Dirn = dirn
	elevator.Requests = r
	elevator.Config = c
}

func setAllLights() {
	//note should be global vaiable
	localFilname := localIP + ".json"
	elevatorName := localIP
	for floor := 0; floor < elevio.NFloors; floor++ {
		for btn := elevio.BHallUp; btn <= elevio.BCab; btn++ {

			checkpoint.JSONsetAllLights(localFilname, elevatorName)
			outputDevice.RequestButtonLight(floor, btn, elevator.Requests[floor][btn])
		}
	}
}

func FsmInitBetweenFloors() {
	dirn := elevio.DirDown
	outputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

//temp testing
/*
func FsmRequestButtonPress(btnFloor int, btn elevio.Button, elevatorName string, filename string) {

	//fmt.Printf("\n\n%s(%d, %s)\n", "FsmRequestButtonPress", btnFloor, elevio.ButtonToString(btn))
	//elev.ElevatorPrint(elevator)

	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		if requests.RequestsShouldClearImmediately(elevator, btnFloor, btn) {
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
		} else {
			//elevator.Requests[btnFloor][btn] = true
			//trenger å sjekke at alt dette er riktig
			fsmUpdateJSONWhenNewOrderOccurs(btnFloor, btn, elevatorName, filename)
			fsmJSONOrderAssigner(filename, elevatorName)
		}

	case elev.EBMoving:
		//elevator.Requests[btnFloor][btn] = true
		//trenger å sjekke at alt dette er riktig
		fsmUpdateJSONWhenNewOrderOccurs(btnFloor, btn, elevatorName, filename)
		fsmJSONOrderAssigner(filename, elevatorName)

	case elev.EBIdle:
		//elevator.Requests[btnFloor][btn] = true
		//trenger å sjekke at alt dette er riktig
		fsmUpdateJSONWhenNewOrderOccurs(btnFloor, btn, elevatorName, filename)
		fsmJSONOrderAssigner(filename, elevatorName)
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)

		case elev.EBMoving:
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevator.Dirn), " in FsmRequestButtonPress")
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights()
	//fmt.Printf("New state: \n")
	//elev.ElevatorPrint(elevator)
}
*/
func FsmFloorArrival(newFloor int, elevatorName string, filename string) {
	//fmt.Printf("\n\n%s(%d)\n", "FsmFloorArrival", newFloor)
	//elev.ElevatorPrint(elevator)
	elevator.CurrentFloor = newFloor
	outputDevice.FloorIndicator(elevator.CurrentFloor)
	//Helt unødvendig med switch her?
	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.RequestsShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			outputDevice.DoorLight(true)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			setAllLights()
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}
	//fmt.Println("New state:")
	//elev.ElevatorPrint(elevator)
}

func FsmDoorTimeout(filename string, elevatorName string) {
	//fmt.Printf("\n\n%s()\n", "FsmDoorTimeout")
	//elev.ElevatorPrint(elevator)
	//Hvorfor switch
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch elevator.CurrentBehaviour {
		case elev.EBDoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)
			setAllLights()

		case elev.EBMoving:
			outputDevice.DoorLight(false)
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevio.DirStop), " in FsmDoorTimeout")
			outputDevice.MotorDirection(elevator.Dirn)
		case elev.EBIdle:
			outputDevice.DoorLight(false)
		}

	}
	//fmt.Println("New State: ")
	//elev.ElevatorPrint(elevator)
}

func FsmObstruction() {
	if !timer.TimerInf {
		timer.TimerStartInf()
		if elevator.CurrentBehaviour == elev.EBIdle {
			outputDevice.DoorLight(true)
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	} else {
		timer.TimerStopInf()
		timer.TimerStart(elevator.Config.DoorOpenDurationS)
	}
}

func FsmMakeCheckpointGo() {
	for {
		checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
		time.Sleep(50 * time.Millisecond)
	}

}

func FsmMakeCheckpoint() {
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func FsmResumeAtLatestCheckpoint(floor int) {
	elevator, _, _ = checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
	setAllLights()
	//fmt.Print(elevator.Dirn)
	if elevator.Dirn != elevio.DirStop && floor == -1 {
		outputDevice.MotorDirection(elevator.Dirn)
	}
	if floor != -1 {
		timer.TimerStart(elev.DoorOpenDurationSConfig)
		outputDevice.DoorLight(true)
	}
}

func FsmLoadLatestCheckpoint() {
	elevator, _, _ = checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
}

// Json fra her
func FsmInitJson(filename string, ElevatorName string) {
	// Gjør endringer på combinedInput her
	print(filename)
	err := os.Remove(filename)
	if err != nil {
		fmt.Println("Feil ved fjerning:", err)
	}
	combinedInput := checkpoint.InitializeCombinedInput(elevator, ElevatorName)

	// If the file was successfully deleted, return nil
	err = checkpoint.SaveCombinedInput(combinedInput, filename)
	if err != nil {
		fmt.Println("Feil ved lagring:", err)
	}
}

func FsmUpdateJSON(elevatorName string, filename string) {
	checkpoint.UpdateJSON(elevator, filename, elevatorName)
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func fsmUpdateJSONWhenNewOrderOccurs(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	checkpoint.UpdateJSONWhenNewOrderOccurs(filename, elevatorName, btnFloor, btn, &elevator)
}

func FsmJSONOrderAssigner(filename string, elevatorName string) {
	checkpoint.JSONOrderAssigner(&elevator, filename, elevatorName)
}

func FsmRequestButtonPressV2(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	if requests.RequestsShouldClearImmediately(elevator, btnFloor, btn) && (elevator.CurrentBehaviour == elev.EBDoorOpen) {
		timer.TimerStart(elevator.Config.DoorOpenDurationS)
	} else {
		//elevator.Requests[btnFloor][btn] = true
		//trenger å sjekke at alt dette er riktig
		fsmUpdateJSONWhenNewOrderOccurs(btnFloor, btn, elevatorName, filename)
		//fsmJSONOrderAssigner(filename, elevatorName)
	}
}

// etter denne func broadcaster vi.
// så assigner vi
// så kaller vi denne
func FsmRequestButtonPressV3(filename string, elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBIdle:
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)

		case elev.EBMoving:
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevator.Dirn), " in FsmRequestButtonPress")
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights()
	//fmt.Printf("New state: \n")
	//elev.ElevatorPrint(elevator)
}

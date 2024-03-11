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

	"github.com/sirupsen/logrus"
)

var elevator elev.Elevator
var outputDevice elevio.ElevOutputDevice
var elevatorName string
var localStateFile string

func init() {
	elevator = elev.ElevatorInit()
	elevatorName, _ = local.GetIP()
	outputDevice = elevio.ElevioGetOutputDevice()

	localStateFile = elevatorName + ".json"

	//Burde dette gå et annet sted?
	setAllLights()
	elevio.RequestDoorOpenLamp(false)
	elevio.RequestStopLamp(false)
}

// BUG: Init and SetElevator crashes when using process pairs
func SetElevator(f int, cb elev.ElevatorBehaviour, dirn elevio.ElevDir, r [elevio.NFloors][elevio.NButtons]bool, c elev.ElevatorConfig) {
	elevator.CurrentFloor = f
	elevator.CurrentBehaviour = cb
	elevator.Dirn = dirn
	elevator.Requests = r
	elevator.Config = c
}

func setAllLights() {
	for floor := 0; floor < elevio.NFloors; floor++ {
		for btn := elevio.BHallUp; btn <= elevio.BCab; btn++ {
			checkpoint.JSONsetAllLights(localStateFile, elevatorName)
			outputDevice.RequestButtonLight(floor, btn, elevator.Requests[floor][btn])
		}
	}
}

func MoveDownToFloor() {
	dirn := elevio.DirDown
	outputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

func FloorArrival(newFloor int, elevatorName string, filename string) {
	logrus.Warn("Arrived at new floor: ", newFloor)
	//elev.ElevatorPrint(elevator)
	elevator.CurrentFloor = newFloor
	outputDevice.FloorIndicator(elevator.CurrentFloor)
	//Helt unødvendig med switch her?
	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.ShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			outputDevice.DoorLight(true)
			elevator = requests.ClearAtCurrentFloor(elevator, filename, elevatorName)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			setAllLights()
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}
	//fmt.Println("New state:")
	//elev.ElevatorPrint(elevator)
}

func DoorTimeout(filename string, elevatorName string) {
	//fmt.Printf("\n\n%s()\n", "DoorTimeout")
	//elev.ElevatorPrint(elevator)
	//Hvorfor switch
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch elevator.CurrentBehaviour {
		case elev.EBDoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator, filename, elevatorName)
			setAllLights()

		case elev.EBMoving:
			outputDevice.DoorLight(false)
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevio.DirStop), " in DoorTimeout")
			outputDevice.MotorDirection(elevator.Dirn)
		case elev.EBIdle:
			outputDevice.DoorLight(false)
		}

	}
	//fmt.Println("New State: ")
	//elev.ElevatorPrint(elevator)
}

func ToggleObstruction() {
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

func MakeCheckpointGo() {
	for {
		checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
		time.Sleep(50 * time.Millisecond)
	}

}

func MakeCheckpoint() {
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func ResumeAtLatestCheckpoint(floor int) {
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

func LoadLatestCheckpoint() {
	elevator, _, _ = checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
}

// Json fra her
func InitJson(filename string, ElevatorName string) {
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

func UpdateJSON(elevatorName string, filename string) {
	checkpoint.UpdateJSON(elevator, filename, elevatorName)
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func RebootJSON(elevatorName string, filename string) {
	checkpoint.RebootJSON(elevator, filename, elevatorName)
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func UpdateJSONOnNewOrder(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	checkpoint.UpdateJSONOnNewOrder(filename, elevatorName, btnFloor, btn, &elevator)
}

func JSONOrderAssigner(filename string, elevatorName string) {
	checkpoint.JSONOrderAssigner(&elevator, filename, elevatorName)
}

func RequestButtonPressV2(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	if requests.ShouldClearImmediately(elevator, btnFloor, btn) && (elevator.CurrentBehaviour == elev.EBDoorOpen) {
		timer.TimerStart(elevator.Config.DoorOpenDurationS)
	} else {
		//elevator.Requests[btnFloor][btn] = true
		//trenger å sjekke at alt dette er riktig
		UpdateJSONOnNewOrder(btnFloor, btn, elevatorName, filename)
		print("funksjonskall funker")
		if btn == elevio.BCab {
			print("hei")
			elevator.Requests[btnFloor][btn] = true
		}
	}
}

// etter denne func broadcaster vi.
// så assigner vi
// så kaller vi denne
func RequestButtonPressV3(filename string, elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBIdle:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator, filename, elevatorName)

		case elev.EBMoving:
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevator.Dirn), " in FsmRequestButtonPress")
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights()
	//fmt.Printf("New state: \n")
	//elev.ElevatorPrint(elevator)
}

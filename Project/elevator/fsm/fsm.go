package fsm

import (
	"elevator/checkpoint"
	"elevator/elev"
	"elevator/elevio"
	"elevator/requests"
	"elevator/timer"
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
	InitJson(localStateFile, elevatorName) 
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

	elevator.CurrentFloor = newFloor
	outputDevice.FloorIndicator(elevator.CurrentFloor)

	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.RequestsShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			outputDevice.DoorLight(true)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)
			timer.Start(elevator.Config.DoorOpenDurationS)
			setAllLights()
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}
}

func DoorTimeout(filename string, elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch elevator.CurrentBehaviour {
		case elev.EBDoorOpen:
			timer.Start(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)
			setAllLights()

		case elev.EBMoving:
			outputDevice.DoorLight(false)
			logrus.Debug("Calling MotorDirection: ", elevio.ElevDirToString(elevio.DirStop))
			outputDevice.MotorDirection(elevator.Dirn)
		case elev.EBIdle:
			outputDevice.DoorLight(false)
		}

	}
}

func ToggleObstruction() {
	if !timer.IsInfinite {
		timer.StartInfiniteTimer()
		if elevator.CurrentBehaviour == elev.EBIdle {
			outputDevice.DoorLight(true)
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	} else {
		timer.StopInfiniteTimer()
		timer.Start(elevator.Config.DoorOpenDurationS)
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
	logrus.Debug("Resuming at last checkpoint at floor: ", floor)
	elevator, _, _ = checkpoint.LoadElevatorCheckpoint(checkpoint.FilenameCheckpoint)
	setAllLights()
	//fmt.Print(elevator.Dirn)
	if elevator.Dirn != elevio.DirStop && floor == -1 {
		outputDevice.MotorDirection(elevator.Dirn)
	}
	if floor != -1 {
		timer.Start(elev.DoorOpenDurationSConfig)
		outputDevice.DoorLight(true)
	}
}

// TODO! JSON code should be a separate module or package. NOT split between fsm and checkpoint packages
func InitJson(filename string, ElevatorName string) {
	// Gjør endringer på combinedInput her
	err := os.Remove(filename)
	if err != nil {
		logrus.Error("Failed to remove file:", err)
	}
	combinedInput := checkpoint.InitializeCombinedInput(elevator, ElevatorName)

	// If the file was successfully deleted, return nil
	err = checkpoint.SaveCombinedInput(combinedInput, filename)
	if err != nil {
		logrus.Error("Failed to save new state:", err)
	}
}

func UpdateJSON(elevatorName string, filename string) {
	checkpoint.UpdateJSON(elevator, filename, elevatorName)
	checkpoint.SetElevatorCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

// TODO: Improve func name
func RebootJSON(elevatorName string, filename string) {
	checkpoint.RebootJSON(elevator, filename, elevatorName)
	checkpoint.SetElevatorCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

// TODO! Could just use checkpoint func directly
func UpdateJSONOnNewOrder(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	checkpoint.UpdateJSONOnNewOrder(filename, elevatorName, btnFloor, btn, &elevator)
}

// TODO! Could just use checkpoint func directly
func JSONOrderAssigner(filename string, elevatorName string) {
	checkpoint.JSONOrderAssigner(&elevator, filename, elevatorName)
}

// TODO! Rewrite func name. This makes no sense
// TODO: Change btn to btnType
func RequestButtonPressV2(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	if requests.RequestsShouldClearImmediately(elevator, btnFloor, btn) && (elevator.CurrentBehaviour == elev.EBDoorOpen) {
		timer.Start(elevator.Config.DoorOpenDurationS)
	} else {
		UpdateJSONOnNewOrder(btnFloor, btn, elevatorName, filename)
		if btn == elevio.BCab {
			elevator.Requests[btnFloor][btn] = true
		}
	}
}

// etter denne func broadcaster vi.
// så assigner vi
// så kaller vi denne
// TODO! Rewrite func name. This makes no sense
func RequestButtonPressV3(filename string, elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBIdle:
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.Start(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)

		case elev.EBMoving:
			logrus.Debug("Calling MotorDirection: ", elevio.ElevDirToString(elevator.Dirn))
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights()
}

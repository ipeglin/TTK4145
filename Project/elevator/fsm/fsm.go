package fsm

import (
	"elevator/checkpoint"
	"elevator/elev"
	"elevator/elevio"
	"elevator/jsonhandler"
	"elevator/requests"
	"elevator/timer"
	"network/local"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var elevator elev.Elevator
var outputDevice elevio.ElevOutputDevice
var nodeIP string

func init() {
	elevator = elev.ElevatorInit()
	nodeIP, _ = local.GetIP()
	outputDevice = elevio.ElevioGetOutputDevice()

	// ? Should this be moved
	SetAllLights()
	elevio.RequestDoorOpenLamp(false)
	elevio.RequestStopLamp(false)
}

func SetAllLights() {
	for floor := 0; floor < elevio.NFloors; floor++ {
		outputDevice.RequestButtonLight(floor, elevio.BCab, elevator.Requests[floor][elevio.BCab])
		if isOffline() || OnlyElevatorOnline(nodeIP) {
			for btn := elevio.BHallUp; btn <= elevio.BCab; btn++ {
				outputDevice.RequestButtonLight(floor, btn, elevator.Requests[floor][btn])
			}
		}
	}
}

func SetConfirmedHallLights(elevatorName string) {
	currentState, _ := jsonhandler.LoadState()
	for floor := 0; floor < elevio.NFloors; floor++ {
		elevio.RequestButtonLight(floor, elevio.BHallUp, currentState.HRAInput.HallRequests[floor][0])
		elevio.RequestButtonLight(floor, elevio.BHallDown, currentState.HRAInput.HallRequests[floor][1])
	}
}

func MoveDownToFloor() {
	dirn := elevio.DirDown
	outputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

func FloorArrival(newFloor int, elevatorName string) {
	logrus.Warn("Arrived at new floor: ", newFloor)

	elevator.CurrentFloor = newFloor
	outputDevice.FloorIndicator(elevator.CurrentFloor)

	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.ShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			outputDevice.DoorLight(true)
			elevator = requests.ClearAtCurrentFloor(elevator, elevatorName)
			timer.Start(elevator.Config.DoorOpenDurationS)
			SetAllLights()
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}

}

func DoorTimeout(elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch elevator.CurrentBehaviour {
		case elev.EBDoorOpen:
			timer.Start(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator, elevatorName)
			SetAllLights()

		case elev.EBMoving:
			outputDevice.DoorLight(false)

			outputDevice.MotorDirection(elevator.Dirn)
		case elev.EBIdle:
			outputDevice.DoorLight(false)
		}
	}
}

func RequestObstruction() {
	if elevator.CurrentBehaviour == elev.EBDoorOpen {
		timer.StartInfiniteTimer()
		jsonhandler.RemoveElevatorsFromJSON([]string{nodeIP})
	}
}

func StopObstruction() {
	timer.StopInfiniteTimer()
	timer.Start(elevator.Config.DoorOpenDurationS)
	HandleStateOnReboot(nodeIP)
}

func CreateCheckpoint() {
	for {
		checkpoint.SetCheckpoint(elevator)
		time.Sleep(50 * time.Millisecond)
	}
}

func ResumeAtLatestCheckpoint(floor int) {
	elevator, _, _ = checkpoint.LoadCheckpoint()
	SetAllLights()

	if elevator.Dirn != elevio.DirStop && floor == -1 {
		outputDevice.MotorDirection(elevator.Dirn)
	}

	if floor != -1 {
		timer.Start(elev.DoorOpenDurationSConfig)
		outputDevice.DoorLight(true)
	}
}

func CreateLocalStateFile(elevatorName string) {
	// TODO: Gjør endringer på elevState her
	err := os.Remove(jsonhandler.StateFile)
	if err != nil {
		logrus.Error("Failed to remove:", err)
	}

	initialElevState := jsonhandler.InitialiseState(elevator, elevatorName)

	// * If the file was successfully deleted, return nil
	err = jsonhandler.SaveState(initialElevState)
	if err != nil {
		logrus.Error("Failed to save checkpoint:", err)
	}
}

// * This was UpdateJSON()
func UpdateElevatorState(elevatorName string) {
	jsonhandler.UpdateJSON(elevator, elevatorName)
	checkpoint.SetCheckpoint(elevator)
}

// * This was RebootJSON()
func HandleStateOnReboot(elevatorName string) {
	jsonhandler.UpdateJSONOnReboot(elevator, elevatorName) // Deprecated: json.RebootJSON()
	checkpoint.SetCheckpoint(elevator)
}

// gir det mening å ha slike oneliners? eller burde vi flytte inn JsonOrderAssignerKoden her?
func AssignOrders(elevatorName string) {
	jsonhandler.JSONOrderAssigner(&elevator, elevatorName)
}

func HandleButtonPress(btnFloor int, btn elevio.Button, elevatorName string) {
	// TODO: Extract the conditions into variables with more informative names
	if requests.ShouldClearImmediately(elevator, btnFloor, btn) && (elevator.CurrentBehaviour == elev.EBDoorOpen) {
		timer.Start(elevator.Config.DoorOpenDurationS)
	} else {
		// TODO: Check if this is correct
		//updateStateOnNewOrder(btnFloor, btn, elevatorName, filename)
		jsonhandler.UpdateJSONOnNewOrder(elevatorName, btnFloor, btn)

		if btn == elevio.BCab {
			elevator.Requests[btnFloor][btn] = true
		}
	}
}

func MoveOnActiveOrders(elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBIdle:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.Start(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator, elevatorName)

		case elev.EBMoving:
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	SetAllLights()
}

func HandleIncomingJSON(localElevatorName string, externalState jsonhandler.ElevatorState, incomingElevatorName string) {
	localState, _ := jsonhandler.LoadState()
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if externalState.Counter.HallRequests[f][i] > localState.Counter.HallRequests[f][i] {
				localState.Counter.HallRequests[f][i] = externalState.Counter.HallRequests[f][i]
				localState.HRAInput.HallRequests[f][i] = externalState.HRAInput.HallRequests[f][i]
			}
			if externalState.Counter.HallRequests[f][i] == localState.Counter.HallRequests[f][i] {
				if localState.HRAInput.HallRequests[f][i] != externalState.HRAInput.HallRequests[f][i] {
					//midliertilig konflikt logikk dersom den ene er true og den andre er false
					//oppstår ved motostop og bostruksjoner etc dersom den har selv claimet en orde som blir utført ila den har motorstop
					//Tenk om dette er beste løsning
					localState.HRAInput.HallRequests[f][i] = false
				}
			}
		}
	}
	if _, exists := externalState.HRAInput.States[incomingElevatorName]; exists {
		if _, exists := localState.HRAInput.States[incomingElevatorName]; !exists {
			localState.HRAInput.States[incomingElevatorName] = externalState.HRAInput.States[incomingElevatorName]
			localState.Counter.States[incomingElevatorName] = externalState.Counter.States[incomingElevatorName]
		} else {
			if externalState.Counter.States[incomingElevatorName] > localState.Counter.States[incomingElevatorName] {
				localState.HRAInput.States[incomingElevatorName] = externalState.HRAInput.States[incomingElevatorName]
				localState.Counter.States[incomingElevatorName] = externalState.Counter.States[incomingElevatorName]
			}
		}
	} else {
		if _, exists := localState.HRAInput.States[incomingElevatorName]; exists {
			delete(localState.HRAInput.States, incomingElevatorName)
			delete(localState.Counter.States, incomingElevatorName)
		}
	}
	if _, exists := externalState.Counter.States[localElevatorName]; exists {
		if externalState.Counter.States[localElevatorName] > localState.Counter.States[localElevatorName] {
			localState.Counter.States[localElevatorName] = externalState.Counter.States[localElevatorName] + 1
		}
	}
	jsonhandler.SaveState(localState)
}

// TODO: Should this go somewehre else?
func worldViewsAlign(localState jsonhandler.ElevatorState, externalState jsonhandler.ElevatorState) bool {
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if externalState.Counter.HallRequests[f][i] != localState.Counter.HallRequests[f][i] {
				return false
			}
		}
	}
	return true
}

func AssignIfWorldViewsAlign(localElevatorName string, externalState jsonhandler.ElevatorState) {
	localState, _ := jsonhandler.LoadState()

	if worldViewsAlign(localState, externalState) {
		jsonhandler.JSONOrderAssigner(&elevator, localElevatorName)
		SetConfirmedHallLights(localElevatorName)
	}
}

// TODO: Maybe IsOnlyNodeOnline() and these functions below dont need to be in fsm?
func OnlyElevatorOnline(localElevatorName string) bool {
	currentState, _ := jsonhandler.LoadState()
	if len(currentState.HRAInput.States) == 1 {
		if _, exists := currentState.HRAInput.States[localElevatorName]; exists {
			return true
		}
	}
	return false
}

func isOffline() bool {
	currentState, _ := jsonhandler.LoadState()
	return len(currentState.HRAInput.States) == 0
}

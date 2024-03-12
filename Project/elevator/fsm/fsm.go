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
var elevatorName string
var localStateFile string

func init() {
	elevator = elev.ElevatorInit()
	elevatorName, _ = local.GetIP()
	outputDevice = elevio.ElevioGetOutputDevice()

	localStateFile = elevatorName + ".json"

	// ? Should this be moved
	setAllLights()
	elevio.RequestDoorOpenLamp(false)
	elevio.RequestStopLamp(false)
}

func setAllLights() {
	for floor := 0; floor < elevio.NFloors; floor++ {
		outputDevice.RequestButtonLight(floor, elevio.BCab, elevator.Requests[floor][elevio.BCab])
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
		if requests.ShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			outputDevice.DoorLight(true)
			elevator = requests.ClearAtCurrentFloor(elevator, filename, elevatorName)
			timer.Start(elevator.Config.DoorOpenDurationS)
			setAllLights()
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}

}

func DoorTimeout(filename string, elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch elevator.CurrentBehaviour {
		case elev.EBDoorOpen:
			timer.Start(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator, filename, elevatorName)
			setAllLights()

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
		jsonhandler.RemoveDysfunctionalElevatorFromJSON(localStateFile, elevatorName)
	}
}

func StopObstruction() {
	timer.StopInfiniteTimer()
	timer.Start(elevator.Config.DoorOpenDurationS)
	HandleStateOnReboot(elevatorName, localStateFile)
}

func CreateCheckpoint() {
	for {
		checkpoint.SetCheckpoint(elevator, checkpoint.CheckpointFilename)
		time.Sleep(50 * time.Millisecond)
	}
}

func ResumeAtLatestCheckpoint(floor int) {
	elevator, _, _ = checkpoint.LoadCheckpoint(checkpoint.CheckpointFilename)
	setAllLights()

	if elevator.Dirn != elevio.DirStop && floor == -1 {
		outputDevice.MotorDirection(elevator.Dirn)
	}
	if floor != -1 {
		timer.Start(elev.DoorOpenDurationSConfig)
		outputDevice.DoorLight(true)
	}
}

func CreateLocalStateFile(filename string, ElevatorName string) {
	// TODO: Gjør endringer på combinedInput her
	err := os.Remove(filename)
	if err != nil {
		logrus.Error("Failed to remove:", err)
	}
	combinedInput := jsonhandler.InitializeCombinedInput(elevator, ElevatorName)

	// * If the file was successfully deleted, return nil
	err = jsonhandler.SaveCombinedInput(combinedInput, filename)
	if err != nil {
		logrus.Error("Failed to save checkpoint:", err)
	}
}

// * This was UpdateJSON()
func UpdateElevatorState(elevatorName string, filename string) {
	jsonhandler.UpdateJSON(elevator, filename, elevatorName)
	checkpoint.SetCheckpoint(elevator, checkpoint.CheckpointFilename)
}

// * This was RebootJSON()
func HandleStateOnReboot(elevatorName string, filename string) {
	jsonhandler.UpdateJSONOnReboot(elevator, filename, elevatorName) // Deprecated: json.RebootJSON()
	checkpoint.SetCheckpoint(elevator, checkpoint.CheckpointFilename)
}

// TODO: Doesn't look like this is used;
func updateStateOnNewOrder(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	jsonhandler.UpdateJSONOnNewOrder(filename, elevatorName, btnFloor, btn)
}

// TODO: Change function name to AssignOrders() or similar
func JSONOrderAssigner(filename string, elevatorName string) {
	jsonhandler.JSONOrderAssigner(&elevator, filename, elevatorName)
}

func HandleButtonPress(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	// TODO: Extract the conditions into variables with more informative names
	if requests.ShouldClearImmediately(elevator, btnFloor, btn) && (elevator.CurrentBehaviour == elev.EBDoorOpen) {
		timer.Start(elevator.Config.DoorOpenDurationS)
	} else {
		// TODO: Check if this is correct
		//updateStateOnNewOrder(btnFloor, btn, elevatorName, filename)
		jsonhandler.UpdateJSONOnNewOrder(filename, elevatorName, btnFloor, btn)

		//TODO: This variable just makes it complicated
		isCabCall := btn == elevio.BCab
		if isCabCall {
			elevator.Requests[btnFloor][btn] = true
		}
	}
}

func MoveOnActiveOrders(filename string, elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBIdle:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.Start(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator, filename, elevatorName)

		case elev.EBMoving:
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights()
}

func HandleIncomingJSON(localFilename string, localElevatorName string, otherCombinedInput jsonhandler.CombinedInput, incomingElevatorName string) {
	localCombinedInput, _ := jsonhandler.LoadCombinedInput(localFilename)
	allValuesEqual := true
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] != localCombinedInput.CyclicCounter.HallRequests[f][i] {
				allValuesEqual = false
				break
			}
		}
	}

	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] > localCombinedInput.CyclicCounter.HallRequests[f][i] {
				localCombinedInput.CyclicCounter.HallRequests[f][i] = otherCombinedInput.CyclicCounter.HallRequests[f][i]
				localCombinedInput.HRAInput.HallRequests[f][i] = otherCombinedInput.HRAInput.HallRequests[f][i]
			}
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] == localCombinedInput.CyclicCounter.HallRequests[f][i] {
				if localCombinedInput.HRAInput.HallRequests[f][i] != otherCombinedInput.HRAInput.HallRequests[f][i] {
					//midliertilig konflikt logikk dersom den ene er true og den andre er false
					//oppstår ved motostop og bostruksjoner etc dersom den har selv claimet en orde som blir utført ila den har motorstop
					//Tenk om dette er beste løsning
					localCombinedInput.HRAInput.HallRequests[f][i] = false
				}
			}
		}
	}
	if _, exists := otherCombinedInput.HRAInput.States[incomingElevatorName]; exists {
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; !exists {
			localCombinedInput.HRAInput.States[incomingElevatorName] = otherCombinedInput.HRAInput.States[incomingElevatorName]
			localCombinedInput.CyclicCounter.States[incomingElevatorName] = otherCombinedInput.CyclicCounter.States[incomingElevatorName]
		} else {
			if otherCombinedInput.CyclicCounter.States[incomingElevatorName] > localCombinedInput.CyclicCounter.States[incomingElevatorName] {
				localCombinedInput.HRAInput.States[incomingElevatorName] = otherCombinedInput.HRAInput.States[incomingElevatorName]
				localCombinedInput.CyclicCounter.States[incomingElevatorName] = otherCombinedInput.CyclicCounter.States[incomingElevatorName]
			}
		}
	} else {
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; exists {
			delete(localCombinedInput.HRAInput.States, incomingElevatorName)
			delete(localCombinedInput.CyclicCounter.States, incomingElevatorName)
		}
	}
	if _, exists := otherCombinedInput.CyclicCounter.States[localElevatorName]; exists {
		if otherCombinedInput.CyclicCounter.States[localElevatorName] > localCombinedInput.CyclicCounter.States[localElevatorName] {
			localCombinedInput.CyclicCounter.States[localElevatorName] = otherCombinedInput.CyclicCounter.States[localElevatorName] + 1
		}
	}
	jsonhandler.SaveCombinedInput(localCombinedInput, localFilename)
	if allValuesEqual {
		jsonhandler.JSONsetAllLights(localFilename, localElevatorName)
		//jsonhandler.JSONOrderAssigner(&elevator, localFilename, localElevatorName)
		jsonhandler.SaveCombinedInput(localCombinedInput, localFilename)
		//oppdater localliste av godetatte hallcalls.
		//fsm.MoveOnActiveOrders(localFilename, localElevatorName) // ! Only have one version
	}
	MoveOnActiveOrders(localFilename, localElevatorName)
	jsonhandler.SaveCombinedInput(localCombinedInput, localFilename)
}

func OnlyElevatorOnlie(localFilename string, localElevatorName string) bool {
	localCombinedInput, _ := jsonhandler.LoadCombinedInput(localFilename)
	if len(localCombinedInput.HRAInput.States) == 1 {
		if _, exists := localCombinedInput.HRAInput.States[localElevatorName]; exists {
			return true
		}
	}
	return false
}

/*
func OflineHandeling(){
	//ikke skru lys av. Finn ut hvor det skjer.
	//Sett alle konfirmed hallcalls til denne lokale heisen
	//move
}
*/

package fsm

//Burde denne filen hetet elevator.go og nåværende elevator.go omdøpes til fsm elns?
import (
	"elevator/checkpoint"
	"elevator/elev"
	"elevator/elevio"
	"elevator/requests"
	"elevator/statehandler"
	"elevator/timer"
	"encoding/json"
	"network/local"
	"os"
	"os/exec"
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
	currentState, _ := statehandler.LoadState()
	isOffline := (len(currentState.HRAInput.States) == 0)

	for floor := 0; floor < elevio.NFloors; floor++ {
		outputDevice.RequestButtonLight(floor, elevio.BCab, elevator.Requests[floor][elevio.BCab])
		if isOffline || statehandler.IsOnlyNodeOnline(nodeIP) {
			for btn := elevio.BHallUp; btn <= elevio.BCab; btn++ {
				outputDevice.RequestButtonLight(floor, btn, elevator.Requests[floor][btn])
			}
		}
	}
}

func SetConfirmedHallLights(elevatorName string) {
	currentState, _ := statehandler.LoadState()
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
		statehandler.RemoveElevatorsFromJSON([]string{nodeIP})
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
	err := os.Remove(statehandler.StateFile)
	if err != nil {
		logrus.Error("Failed to remove:", err)
	}

	initialElevState := statehandler.InitialiseState(elevator, elevatorName)

	// * If the file was successfully deleted, return nil
	err = statehandler.SaveState(initialElevState)
	if err != nil {
		logrus.Error("Failed to save checkpoint:", err)
	}
}

func UpdateElevatorState(elevatorName string) {
	statehandler.UpdateState(elevator, elevatorName)
	checkpoint.SetCheckpoint(elevator)
}

func HandleStateOnReboot(elevatorName string) {
	statehandler.UpdateStateOnReboot(elevator, elevatorName)
	checkpoint.SetCheckpoint(elevator)
}

func AssignOrders(elevatorName string) {
	state, err := statehandler.LoadState()
	if err != nil {
		logrus.Debugf("Failed to load combined input: %v\n", err)
		return
	}

	if len(state.HRAInput.States) == 0 {
		logrus.Debug("HRAInput.States is empty, skipping order assignment")
		return
	}

	jsonBytes, err := json.Marshal(state.HRAInput)
	if err != nil {
		logrus.Debugf("Failed to marshal HRAInput: %v\n", err)
		return
	}

	ret, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		logrus.Debugf("exec.Command error: %v\nOutput: %s\n", err, string(ret))
		return
	}

	output := make(map[string][][2]bool)
	if err := json.Unmarshal(ret, &output); err != nil {
		logrus.Debugf("json.Unmarshal error: %v\n", err)
		return
	}

	for floor := 0; floor < elevio.NFloors; floor++ {
		if orders, ok := output[elevatorName]; ok && floor < len(orders) {
			elevator.Requests[floor][elevio.BHallUp] = orders[floor][0]
			elevator.Requests[floor][elevio.BHallDown] = orders[floor][1]
		}
	}
}

func AssignIfWorldViewsAlign(localElevatorName string, externalState statehandler.ElevatorState) {
	localState, _ := statehandler.LoadState()
	WView := true
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if externalState.Counter.HallRequests[f][i] != localState.Counter.HallRequests[f][i] {
				WView = false
			}
		}
	}
	if WView {
		AssignOrders(localElevatorName)
		SetConfirmedHallLights(localElevatorName)
	}
}

func HandleButtonPress(btnFloor int, btn elevio.Button, elevatorName string) {
	// TODO: Extract the conditions into variables with more informative names
	if requests.ShouldClearImmediately(elevator, btnFloor, btn) && (elevator.CurrentBehaviour == elev.EBDoorOpen) {
		timer.Start(elevator.Config.DoorOpenDurationS)
	} else {
		statehandler.UpdateStateOnNewOrder(elevatorName, btnFloor, btn)

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

// Todo:: functions below dont need to be in fsm?

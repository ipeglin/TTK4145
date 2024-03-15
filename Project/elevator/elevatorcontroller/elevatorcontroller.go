package elevatorcontroller

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
var nodeIP string

func init() {
	elevator = elev.ElevatorInit()
	nodeIP, _ = local.GetIP()
	SetAllLights()
	elevio.RequestDoorOpenLamp(false)
	elevio.RequestStopLamp(false)
}

func GetElevator() elev.Elevator {
	return elevator
}

func SetAllLights() {
	currentState, _ := statehandler.LoadState()
	isOffline := (len(currentState.HRAInput.States) == 0)

	for floor := 0; floor < elevio.NFloors; floor++ {
		elevio.OutputDevice.RequestButtonLight(floor, elevio.BCab, elevator.Requests[floor][elevio.BCab])
		if isOffline || statehandler.IsOnlyNodeOnline(nodeIP) {
			for btn := elevio.BHallUp; btn <= elevio.BCab; btn++ {
				elevio.OutputDevice.RequestButtonLight(floor, btn, elevator.Requests[floor][btn])
			}
		}
	}
}

func SetConfirmedHallLights(elevatorName string) {
	currentState, _ := statehandler.LoadState()
	for floor := 0; floor < elevio.NFloors; floor++ {
		for btn := elevio.BHallUp; btn < elevio.BCab; btn++ {
			elevio.RequestButtonLight(floor, btn, currentState.HRAInput.HallRequests[floor][btn])
		}
	}
}

func MoveDownToFloor() {
	dirn := elevio.DirDown
	elevio.OutputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

func FloorArrival(newFloor int, elevatorName string) {
	logrus.Warn("Arrived at new floor: ", newFloor)

	elevator.CurrentFloor = newFloor
	elevio.OutputDevice.FloorIndicator(elevator.CurrentFloor)

	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.ShouldStop(elevator) {
			elevio.OutputDevice.MotorDirection(elevio.DirStop)
			elevio.OutputDevice.DoorLight(true)
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
			elevio.OutputDevice.DoorLight(false)

			elevio.OutputDevice.MotorDirection(elevator.Dirn)
		case elev.EBIdle:
			elevio.OutputDevice.DoorLight(false)
		}
	}
}

func RequestObstruction() {
	if elevator.CurrentBehaviour == elev.EBDoorOpen {
		timer.StartInfiniteTimer()
		statehandler.RemoveElevatorsFromState([]string{nodeIP})
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
		elevio.OutputDevice.MotorDirection(elevator.Dirn)
	}

	if floor != -1 {
		timer.Start(elev.DoorOpenDurationSConfig)
		elevio.OutputDevice.DoorLight(true)
		elevator.CurrentBehaviour = elev.EBDoorOpen
	}
}

func CreateLocalStateFile(elevatorName string) {
	err := os.Remove(statehandler.StateFile)
	if err != nil {
		logrus.Error("Failed to remove:", err)
	}

	initialElevState := statehandler.InitialiseState(elevator, elevatorName)
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
			for btn := elevio.BHallUp; btn < elevio.BCab; btn++ {
				elevator.Requests[floor][btn] = orders[floor][btn]
			}
		}
	}
}

func AssignIfWorldViewsAlign(localElevatorName string, externalState statehandler.ElevatorState) {
	localState, _ := statehandler.LoadState()

	if isOrderStatesEqual(localState, externalState) {
		AssignOrders(localElevatorName)
		SetConfirmedHallLights(localElevatorName)
	}
}

func isOrderStatesEqual(state statehandler.ElevatorState, externalState statehandler.ElevatorState) bool {
	for f := 0; f < elevio.NFloors; f++ {
		for btn := elevio.BHallUp; btn < elevio.BCab; btn++ {
			if externalState.Counter.HallRequests[f][btn] != state.Counter.HallRequests[f][btn] {
				return false
			}
		}
	}

	return true
}

func HandleButtonPress(btnFloor int, btn elevio.Button, elevatorName string) {
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
			elevio.OutputDevice.DoorLight(true)
			timer.Start(elevator.Config.DoorOpenDurationS)
			elevator = requests.ClearAtCurrentFloor(elevator, elevatorName)
		case elev.EBMoving:
			elevio.OutputDevice.MotorDirection(elevator.Dirn)
		}
	}
	SetAllLights()
}

func MontitorMotorActivity(receiver chan<- bool) {
	timerActive := true
	timerEndTimer := timer.GetCurrentTimeAsFloat() + elev.MotorTimeoutS
	v := elevio.RequestFloor()
	for {
		time.Sleep(elevio.PollRateMS * time.Millisecond)
		if v != -1 && (elevator.CurrentBehaviour != elev.EBMoving) || v != elevio.RequestFloor() {
			timerEndTimer = timer.GetCurrentTimeAsFloat() + elev.MotorTimeoutS
			if !timerActive {
				timerActive = true
				receiver <- true
			}
		} else {
			if timer.GetCurrentTimeAsFloat() > timerEndTimer {
				if timerActive {
					timerActive = false
					receiver <- false
				}
			}
		}
		v = elevio.RequestFloor()
	}
}

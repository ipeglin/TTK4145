package elevator

import (
	"elevator/elevio"
	"elevator/fsm"
	"elevator/statehandler"
	"elevator/timer"
	"time"

	"github.com/sirupsen/logrus"
)

func Init(elevatorName string, isPrimaryProcess bool) {
	logrus.Info("Elevator module initiated with name ", elevatorName)
	if isPrimaryProcess {
		if elevio.InputDevice.FloorSensor() == -1 {
			logrus.Info("Elevator initialised between floors")
			fsm.MoveDownToFloor()
		}
		fsm.CreateLocalStateFile(elevatorName)
	} else {
		floor := elevio.InputDevice.FloorSensor()
		fsm.ResumeAtLatestCheckpoint(floor)
	}

	buttons := make(chan elevio.ButtonEvent)
	floors := make(chan int)
	obst := make(chan bool)
	motorActivity := make(chan bool)

	go elevio.PollButtons(buttons)
	go elevio.PollFloorSensor(floors)
	go elevio.PollObstructionSwitch(obst)
	go elevio.MontitorMotorActivity(motorActivity)
	go fsm.CreateCheckpoint()

	var obstructed bool = false
	for {
		select {
		case obstructed = <-obst:
			logrus.Warn("Obstruction state changed: ", obstructed)
			if obstructed {
				logrus.Debug("New obstruction detected: ", obstructed)
				fsm.RequestObstruction()
			} else {
				fsm.StopObstruction()
			}

		case motorActive := <-motorActivity:
			logrus.Warn("MotorActive state changed: ", motorActive)
			if !motorActive {
				statehandler.RemoveElevatorsFromState([]string{elevatorName})
			} else {
				fsm.HandleStateOnReboot(elevatorName)
				fsm.MoveOnActiveOrders(elevatorName)
			}

		case btnEvent := <-buttons:
			logrus.Debug("Button press detected: ", btnEvent)
			fsm.UpdateElevatorState(elevatorName)
			fsm.HandleButtonPress(btnEvent.Floor, btnEvent.Button, elevatorName)
			if statehandler.IsOnlyNodeOnline(elevatorName) {
				fsm.AssignOrders(elevatorName)
			}
			fsm.MoveOnActiveOrders(elevatorName)
			fsm.UpdateElevatorState(elevatorName)

		case floor := <-floors:
			logrus.Debug("Floor sensor triggered: ", floor)
			fsm.FloorArrival(floor, elevatorName)
			fsm.UpdateElevatorState(elevatorName)
			if statehandler.IsOnlyNodeOnline(elevatorName) {
				fsm.AssignOrders(elevatorName)
			}
			if obstructed {
				fsm.RequestObstruction()
			}

		default:
			if timer.TimedOut() {
				logrus.Debug("Elevator timeout")
				fsm.UpdateElevatorState(elevatorName)
				timer.Stop()
				fsm.DoorTimeout(elevatorName)
				fsm.UpdateElevatorState(elevatorName)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

}

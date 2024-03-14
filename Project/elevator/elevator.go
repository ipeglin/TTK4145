package elevator

import (
	"elevator/elevatorcontroller"
	"elevator/elevio"
	"elevator/motoractivity"
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
			elevatorcontroller.MoveDownToFloor()
		}
		elevatorcontroller.CreateLocalStateFile(elevatorName)
	} else {
		floor := elevio.InputDevice.FloorSensor()
		elevatorcontroller.ResumeAtLatestCheckpoint(floor)
	}

	buttons := make(chan elevio.ButtonEvent)
	floors := make(chan int)
	obst := make(chan bool)
	motorActivity := make(chan bool)

	go elevio.PollButtons(buttons)
	go elevio.PollFloorSensor(floors)
	go elevio.PollObstructionSwitch(obst)
	go motoractivity.MontitorMotorActivity(motorActivity)
	go elevatorcontroller.CreateCheckpoint()

	var obstructed bool = false
	for {
		select {
		case obstructed = <-obst:
			logrus.Warn("Obstruction state changed: ", obstructed)
			if obstructed {
				logrus.Debug("New obstruction detected: ", obstructed)
				elevatorcontroller.RequestObstruction()
			} else {
				elevatorcontroller.StopObstruction()
			}

		case motorActive := <-motorActivity:
			logrus.Warn("MotorActive state changed: ", motorActive)
			if !motorActive {
				statehandler.RemoveElevatorsFromState([]string{elevatorName})
			} else {
				elevatorcontroller.HandleStateOnReboot(elevatorName)
				elevatorcontroller.MoveOnActiveOrders(elevatorName)
			}

		case btnEvent := <-buttons:
			print("button pressed")
			logrus.Debug("Button press detected: ", btnEvent)
			elevatorcontroller.UpdateElevatorState(elevatorName)
			elevatorcontroller.HandleButtonPress(btnEvent.Floor, btnEvent.Button, elevatorName)
			if statehandler.IsOnlyNodeOnline(elevatorName) {
				elevatorcontroller.AssignOrders(elevatorName)
			}
			elevatorcontroller.MoveOnActiveOrders(elevatorName)
			elevatorcontroller.UpdateElevatorState(elevatorName)

		case floor := <-floors:
			logrus.Debug("Floor sensor triggered: ", floor)
			elevatorcontroller.FloorArrival(floor, elevatorName)
			elevatorcontroller.UpdateElevatorState(elevatorName)
			if statehandler.IsOnlyNodeOnline(elevatorName) {
				elevatorcontroller.AssignOrders(elevatorName)
			}
			if obstructed {
				elevatorcontroller.RequestObstruction()
			}

		default:
			if timer.TimedOut() {
				logrus.Debug("Elevator timeout")
				elevatorcontroller.UpdateElevatorState(elevatorName)
				timer.Stop()
				elevatorcontroller.DoorTimeout(elevatorName)
				elevatorcontroller.UpdateElevatorState(elevatorName)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

}

package elevator

import (
	"elevator/elevio"
	"elevator/fsm"
	"elevator/jsonhandler"
	"elevator/timer"
	"time"

	"github.com/sirupsen/logrus"
)

func Init(elevatorName string, isPrimaryProcess bool) {
	logrus.Info("Elevator module initiated with name ", elevatorName)
	elevatorStateFile := elevatorName + ".json"
	if isPrimaryProcess {
		if elevio.InputDevice.FloorSensor() == -1 {
			logrus.Info("Elevator initialised between floors")
			fsm.MoveDownToFloor()
		}
		fsm.CreateLocalStateFile(elevatorStateFile, elevatorName)
	} else {
		floor := elevio.InputDevice.FloorSensor()
		fsm.ResumeAtLatestCheckpoint(floor)
	}

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_motorActivity := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevio.MontitorMotorActivity(drv_motorActivity, 3.0)
	go fsm.CreateCheckpoint()

	// initial hinderance states
	var obst bool = false
	for {
		select {
		case obst = <-drv_obstr:
			logrus.Warn("Obstruction state changed: ", obst)
			if obst {
				logrus.Debug("New obstruction detected: ", obst)
				fsm.RequestObstruction()
			} else {
				fsm.StopObstruction()
			}

		case motorActive := <-drv_motorActivity:
			logrus.Warn("MotorActive state changed: ", motorActive)
			if !motorActive {
				jsonhandler.RemoveElevatorsFromJSON([]string{elevatorName}, elevatorStateFile)
			} else {
				fsm.HandleStateOnReboot(elevatorName, elevatorStateFile)
			}

		case btnEvent := <-drv_buttons:
			logrus.Debug("Button press detected: ", btnEvent)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
			fsm.HandleButtonPress(btnEvent.Floor, btnEvent.Button, elevatorName, elevatorStateFile)

			if fsm.OnlyElevatorOnline(elevatorStateFile, elevatorName) {
				fsm.AssignOrders(elevatorStateFile, elevatorName)
			}

			fsm.MoveOnActiveOrders(elevatorStateFile, elevatorName)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)

		case floor := <-drv_floors:
			logrus.Debug("Floor sensor triggered: ", floor)
			fsm.FloorArrival(floor, elevatorName, elevatorStateFile)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)

			if fsm.OnlyElevatorOnline(elevatorStateFile, elevatorName) {
				fsm.AssignOrders(elevatorStateFile, elevatorName)
			}
			if obst {
				fsm.RequestObstruction()
			}

		default:
			if timer.TimedOut() {
				logrus.Debug("Elevator timeout")
				fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
				timer.Stop()
				fsm.DoorTimeout(elevatorStateFile, elevatorName)
				fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

}

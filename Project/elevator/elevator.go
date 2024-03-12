package elevator

import (
	"elevator/driver/hwelevio"
	"elevator/elevio"
	"elevator/fsm"
	"elevator/jsonhandler"
	"elevator/timer"
	"time"

	"github.com/sirupsen/logrus"
)

func Init(elevatorName string, isPrimaryProcess bool) {
	logrus.Info("Elevator module initiated with name ", elevatorName)

	hwelevio.Init(elevio.Addr, elevio.NFloors)
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
	// TODO: Add channels for direction and behaviour

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevio.MontitorMotorActivity(drv_motorActivity, 3.0)
	go fsm.CreateCheckpoint()
	// TODO: Add polling for direction and behaviour

	// initial hinderance states
	var obst bool = false
	for {
		select {
		case obst = <-drv_obstr:
			logrus.Warn("Obstruction state changed: ", obst)
			if obst { // If obstruction detected and it's a new obstruction
				logrus.Debug("New obstruction detected: ", obst)
				fsm.RequestObstruction()
			} else {
				fsm.StopObstruction()
			}

		case motorActive := <-drv_motorActivity:
			logrus.Warn("Immobile state changed: ", motorActive)
			if !motorActive {
				// BUG: THis occurs very late
				jsonhandler.RemoveDysfunctionalElevatorFromJSON(elevatorStateFile, elevatorName)
				//we need to remove the request// clear them if we dont want to comlete orders twice.
				//it is up to uss and we have functionality to do so
			} else {
				fsm.HandleStateOnReboot(elevatorName, elevatorStateFile)
				//lurer på om vi må ha en movebutton her men idk

				//fsm.MoveOnActiveOrders(elevatorStateFile, elevatorName)
				//fsm.JSONOrderAssigner(elevatorStateFile, elevatorName)
			}
		//TODO: Alle UpdateElevatorState should be in the fsm functions beeing called
		case btnEvent := <-drv_buttons:
			logrus.Debug("Button press detected: ", btnEvent)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
			fsm.HandleButtonPress(btnEvent.Floor, btnEvent.Button, elevatorName, elevatorStateFile)
			if fsm.OnlyElevatorOnlie(elevatorStateFile, elevatorName) {
				print("jeg er eneste onlibe")
				fsm.JSONOrderAssigner(elevatorStateFile, elevatorName)
				jsonhandler.JSONsetAllLights(elevatorStateFile, elevatorName)
			}
			fsm.MoveOnActiveOrders(elevatorStateFile, elevatorName)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
		//TODO: Alle UpdateElevatorState should be in the fsm functions beeing called
		case floor := <-drv_floors:
			logrus.Debug("Floor sensor triggered: ", floor)
			fsm.FloorArrival(floor, elevatorName, elevatorStateFile)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
			if fsm.OnlyElevatorOnlie(elevatorStateFile, elevatorName) {
				//fsm.JSONOrderAssigner(elevatorStateFile, elevatorName)
				jsonhandler.JSONsetAllLights(elevatorStateFile, elevatorName)
			}
			if obst {
				fsm.RequestObstruction()
			}

		default:
			if timer.TimedOut() { // Check for timeout only if no obstruction
				logrus.Debug("Elevator timeout")
				fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
				timer.Stop()
				fsm.DoorTimeout(elevatorStateFile, elevatorName)
				fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
			}
			time.Sleep(50 * time.Millisecond)
		}
		/// we need a case for each time a state updates.
	}

}

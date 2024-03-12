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
			// elevator initialised between floors
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
	drv_obstr_immob := make(chan bool)
	drv_stop := make(chan bool)
	drv_motorActivity := make(chan bool)
	//immob := make(chan bool)
	// TODO: Add channels for direction and behaviour

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevio.MontitorMotorActivity(drv_motorActivity, 3.0)
	//go immobility.Immobility(drv_obstr_immob, drv_motorActivity, immob)
	go fsm.CreateCheckpoint()
	// TODO: Add polling for direction and behaviour

	// initial hinderance states
	var obst bool = false
	var immobile bool = false

	for {
		select {
		case drv_obst := <-drv_obstr:
			logrus.Warn("Obstruction state changed: ", drv_obst)
			drv_obstr_immob <- drv_obst
			if drv_obst == !obst { // If obstruction detected and it's a new obstruction
				logrus.Debug("New obstruction detected: ", drv_obst)
				fsm.ToggleObstruction()
			} else {
				fsm.HandleStateOnReboot(elevatorName, elevatorStateFile)
			}
			obst = drv_obst

		case motorStop := <-drv_motorActivity:
			logrus.Warn("Immobile state changed: ", immobile)
			if motorStop {
				// BUG: THis occurs very late
				jsonhandler.RemoveDysfunctionalElevatorFromJSON(elevatorStateFile, elevatorName)
				//we need to remove the request// clear them if we dont want to comlete orders twice.
				//it is up to uss and we have functionality to do so
			} else {
				fsm.HandleStateOnReboot(elevatorName, elevatorStateFile)
				//fsm.MoveOnActiveOrders(elevatorStateFile, elevatorName)
				//fsm.JSONOrderAssigner(elevatorStateFile, elevatorName)
			}

		case btnEvent := <-drv_buttons:
			logrus.Debug("Button press detected: ", btnEvent)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
			//trenger ikke være her. assign kun ved innkomende mld da heis offline ikke skal assigne
			//print("hjelp noe må funke")
			fsm.HandleButtonPress(btnEvent.Floor, btnEvent.Button, elevatorName, elevatorStateFile)
			//fsm.JSONOrderAssigner(elevatorStateFile, elevatorName)

			fsm.MoveOnActiveOrders(elevatorStateFile, elevatorName)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)

		case floor := <-drv_floors:
			logrus.Debug("Floor sensor triggered: ", floor)
			fsm.FloorArrival(floor, elevatorName, elevatorStateFile)
			fsm.UpdateElevatorState(elevatorName, elevatorStateFile)
			if obst {
				jsonhandler.RemoveDysfunctionalElevatorFromJSON(elevatorStateFile, elevatorName)
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

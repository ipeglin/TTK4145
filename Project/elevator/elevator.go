package elevator

import (
	"elevator/checkpoint"
	"elevator/driver/hwelevio"
	"elevator/elevio"
	"elevator/fsm"
	"elevator/immobility"
	"elevator/timer"

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
		fsm.FsmInitJson(elevatorStateFile, elevatorName)
	} else {
		floor := elevio.InputDevice.FloorSensor()
		fsm.FsmResumeAtLatestCheckpoint(floor)
	}

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_obstr_immob := make(chan bool)
	drv_stop := make(chan bool)
	drv_motorActivity := make(chan bool)
	immob := make(chan bool)
	// TODO: Add channels for direction and behaviour

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevio.MontitorMotorActivity(drv_motorActivity, 3.0)
	go immobility.Immobility(drv_obstr_immob, drv_motorActivity, immob)
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
				fsm.FsmObstruction()
			}
			obst = drv_obst

		case immobile = <-immob:
			logrus.Warn("Immobile state changed: ", immobile)
			if immobile {
				// BUG: THis occurs very late
				checkpoint.RemoveDysfunctionalElevatorFromJSON(elevatorStateFile, elevatorName)
			} else {
				fsm.FsmUpdateJSON(elevatorName, elevatorStateFile)
			}

		case btnEvent := <-drv_buttons:
			logrus.Debug("Button press detected: ", btnEvent)
			fsm.FsmUpdateJSON(elevatorName, elevatorStateFile)
			//trenger ikke være her. assign kun ved innkomende mld da heis offline ikke skal assigne
			fsm.FsmRequestButtonPressV2(btnEvent.Floor, btnEvent.Button, elevatorName, elevatorStateFile)
			//fsm.FsmJSONOrderAssigner(elevatorStateFile, elevatorName)
			//fsm.FsmRequestButtonPressV3(elevatorStateFile, elevatorName)
			fsm.FsmUpdateJSON(elevatorName, elevatorStateFile)

		case floor := <-drv_floors:
			logrus.Debug("Floor sensor triggered: ", floor)
			fsm.FloorArrival(floor, elevatorName, elevatorStateFile)
			fsm.FsmUpdateJSON(elevatorName, elevatorStateFile)
			fsm.FsmMakeCheckpoint()

		default:
			if timer.TimerTimedOut() { // Check for timeout only if no obstruction
				logrus.Debug("Elevator timeout")
				fsm.FsmUpdateJSON(elevatorName, elevatorStateFile)
				timer.TimerStop()
				fsm.FsmDoorTimeout(elevatorStateFile, elevatorName)
				fsm.FsmUpdateJSON(elevatorName, elevatorStateFile)
			}
			fsm.FsmMakeCheckpoint()
		}
		/// we need a case for each time a state updates.
	}

}

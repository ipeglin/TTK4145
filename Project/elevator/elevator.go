package elevator

import (
	"elevator/checkpoint"
	"elevator/driver/hwelevio"
	"elevator/elevio"
	"elevator/fsm"
	"elevator/immobility"
	"elevator/timer"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func Init(localIP string, firstProcess bool) {
	elevatorName := localIP
	logrus.Info("Elevator module initiated with name ", localIP)

	hwelevio.Init(elevio.Addr, elevio.NFloors)
	filename := elevatorName + ".json"
	if firstProcess {
		if elevio.InputDevice.FloorSensor() == -1 {
			fsm.FsmInitBetweenFloors()
		}
		fsm.FsmInitJson(filename, elevatorName)
	} else {
		floor := elevio.InputDevice.FloorSensor()
		fmt.Print("\n Resuming")
		fsm.FsmResumeAtLatestCheckpoint(floor)
	}

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_obstr_immob := make(chan bool)
	drv_stop := make(chan bool)
	drv_motorActivity := make(chan bool)
	immob := make(chan bool)
	//todo
	//drv_direction :=make(chan int)
	//drv_behaviour

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevio.MontitorMotorActivity(drv_motorActivity, 3.0)
	go immobility.Immobility(drv_obstr_immob, drv_motorActivity, immob)
	go fsm.FsmMakeCheckpointGo()
	//go elvio.PollDirection(drv_direction)
	//go elvio.PollBehaviour(drv_behaviour)

	var obst bool = false
	var immobile bool = false
	for {
		select {
		case drv_obst := <-drv_obstr:
			logrus.Warn("Obstruction state changed: ", drv_obst)
			drv_obstr_immob <- drv_obst
			if drv_obst == !obst { // If obstruction detected and it's a new obstruction
				fsm.FsmObstruction()
			}
			obst = drv_obst

		case immobile = <-immob:
			if immobile {
				//dette skjer veldig sent
				checkpoint.RemoveDysfunctionalElevatorFromJSON(filename, elevatorName)
				logrus.Warn("Immobile state changed: ", immobile)
			} else {
				//fra imobil til mobil
				fsm.FsmUpdateJSON(elevatorName, filename)
			}

		case btnEvent := <-drv_buttons:
			logrus.Debug("Button press detected: ", btnEvent)
			fsm.FsmUpdateJSON(elevatorName, filename)
			//trenger ikke være her. assign kun ved innkomende mld da heis offline ikke skal assigne
			fsm.FsmRequestButtonPressV2(btnEvent.Floor, btnEvent.Button, elevatorName, filename)
			fsm.FsmJSONOrderAssigner(filename, elevatorName)
			fsm.FsmRequestButtonPressV3(filename, elevatorName)
			fsm.FsmUpdateJSON(elevatorName, filename)

		case floor := <-drv_floors:
			logrus.Debug("Floor sensor triggered: ", floor)
			fsm.FsmFloorArrival(floor, elevatorName, filename)
			fsm.FsmUpdateJSON(elevatorName, filename)
			fsm.FsmMakeCheckpoint()

		default:
			if timer.TimerTimedOut() { // Check for timeout only if no obstruction
				logrus.Debug("Elevator timeout")
				fsm.FsmUpdateJSON(elevatorName, filename)
				timer.TimerStop()
				fsm.FsmDoorTimeout(filename, elevatorName)
				fsm.FsmUpdateJSON(elevatorName, filename)
			}
			time.Sleep(50 * time.Millisecond)
		}
		/// we need a case for each time a state updates.
	}

}

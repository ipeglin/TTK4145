package elevator

import (
	"elevator/driver/hwelevio"
	"elevator/elevio"
	"elevator/fsm"
	"elevator/immobility"
	"elevator/timer"
	"fmt"
)

func Init(localIP string, firstProcess bool) {
	fmt.Println("Started!")
	elevatorName := localIP

	hwelevio.Init(elevio.Addr, elevio.NFloors)
	filename := elevatorName + ".json"
	if firstProcess {
		if elevio.InputDevice.FloorSensor() == -1 {
			fsm.FsmInitBetweenFloors()
		}
		fsm.FsmInitJson(filename, elevatorName)
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
	//todo
	//drv_direction :=make(chan int)
	//drv_behaviour

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevio.MontitorMotorActivity(drv_motorActivity, 3.0)
	go immobility.Immobility(drv_obstr_immob, drv_motorActivity, immob)
	//go elvio.PollDirection(drv_direction)
	//go elvio.PollBehaviour(drv_behaviour)

	var obst bool = false
	var immobile bool = false
	for {
		select {
		case drv_obst := <-drv_obstr:
			//fmt.Print(("obst"))
			drv_obstr_immob <- drv_obst
			if drv_obst == !obst { // If obstruction detected and it's a new obstruction
				fsm.FsmObstruction()
			}
			obst = drv_obst

		case immobile = <-immob:
			if immobile {
				//TODO
				fmt.Print("Immobile")
			}
			fmt.Print(immobile)

		case btnEvent := <-drv_buttons:
			fsm.FsmUpdateJSON(elevatorName, filename)
			fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button, elevatorName, filename)
			fsm.FsmUpdateJSON(elevatorName, filename)

		case floor := <-drv_floors:
			//fmt.Print("Arrived")
			fsm.FsmFloorArrival(floor, elevatorName, filename)
			fsm.FsmUpdateJSON(elevatorName, filename)

		default:
			if timer.TimerTimedOut() && !obst { // Check for timeout only if no obstruction
				fsm.FsmUpdateJSON(elevatorName, filename)
				timer.TimerStop()
				fsm.FsmDoorTimeout()
				fsm.FsmUpdateJSON(elevatorName, filename)
			}
			fsm.FsmMakeCheckpoint()
		}
		/// we need a case for each time a state updates.
	}

}

package elevator

import (
	"elevator/driver/hwelevio"
	"elevator/elevio"
	"elevator/fsm"
	"elevator/timer"
	"fmt"
)

func Init(localIP string, firstProcess) {
	fmt.Println("Started!")
	elevatorName := localIP

	hwelevio.Init(elevio.Addr, elevio.NFloors)

	if firstProcess {
		if elevio.InputDevice.FloorSensor() == -1 {
			fsm.FsmInitBetweenFloors()
		}
		filename := elevatorName + ".json"
		fsm.FsmInitJson(filename, elevatorName)
	}
	

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	//todo
	//drv_direction :=make(chan int)
	//drv_behaviour

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	//go elvio.PollDirection(drv_direction)
	//go elvio.PollBehaviour(drv_behaviour)

	var obst bool = false
	var stop bool = false
	for {
		select {
		case drv_obst := <-drv_obstr:
			//fmt.Print(("obst"))
			if drv_obst == !obst { // If obstruction detected and it's a new obstruction
				fsm.FsmObstruction()
			}
			obst = drv_obst

		case drv_stp := <-drv_stop:
			//fmt.Print("Stopp")
			if drv_stp != stop { // If there's a change in the stop signal
				stop = drv_stp
				//fsm.FsmStop(stop)
			}

		case btnEvent := <-drv_buttons:
			if !stop { // Process button presses only if not stopped
				fsm.FsmUpdateJSON(elevatorName,filename)
				fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button, elevatorName, filename)
				fsm.FsmUpdateJSON(elevatorName, filename)
			}

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

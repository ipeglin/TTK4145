package elevator

import (
	"elevator/driver/hwelevio"
	"elevator/elevio"
	"elevator/fsm"
	"elevator/timer"
	"fmt"
)

func Init() {

	fmt.Println("Started!")
	hwelevio.Init(elevio.Addr, elevio.NFloors)

	if elevio.InputDevice.FloorSensor() == -1 {
		fsm.FsmInitBetweenFloors()
	}
	fsm.FsmInitCyclicCounter()

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
			fmt.Print(("obst"))
			if drv_obst == !obst { // If obstruction detected and it's a new obstruction
				fsm.FsmObstruction()
			}
			obst = drv_obst

		case drv_stp := <-drv_stop:
			fmt.Print("Stopp")
			if drv_stp != stop { // If there's a change in the stop signal
				stop = drv_stp
				//fsm.FsmStop(stop)
			}

		case btnEvent := <-drv_buttons:
			if !stop { // Process button presses only if not stopped
				//finn ut hvilken knap trykkes og oppdater cylick counter

				//her bør executable calles som skal oppdatere 
				//så oppdater elevator med ny data


				fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button)


				fsm.FsmUpdataJSONOnbtnEvent()
				fsm.FsmUpdateCylickCounterButtonPressed(btnEvent.Floor, btnEvent.Button)
			}

		case floor := <-drv_floors:
			print("halla")
			fmt.Print("Arrived")
			fsm.FsmFloorArrival(floor)
			fsm.FsmUpdateLocalElevatorToJSON()
			fsm.FsmUpdateCylickCounterNewFloor()

		default:
			if timer.TimerTimedOut() && !obst { // Check for timeout only if no obstruction
				timer.TimerStop()
				fsm.FsmDoorTimeout()
			}
		}
		/// we need a case for each time a state updates. 
	}

}


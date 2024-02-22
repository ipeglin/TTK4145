package main

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
	"heislab/Elevator/elevio"
	"heislab/Elevator/fsm"
	"heislab/Elevator/timer"
)

func mainLogic() {
	fmt.Println("Started!")
	hwelevio.Init(elevio.Addr, elevio.NFloors)

	if elevio.InputDevice.FloorSensor() == -1 {
		fsm.FsmInitBetweenFloors()
	}

	drv_buttons := make(chan hwelevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go hwelevio.PollButtons(drv_buttons)
	go hwelevio.PollFloorSensor(drv_floors)
	go hwelevio.PollObstructionSwitch(drv_obstr)
	go hwelevio.PollStopButton(drv_stop)

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
				fsm.FsmStop(stop)
			}

		case btnEvent := <-drv_buttons:
			if !stop { // Process button presses only if not stopped
				fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button)
			}

		case floor := <-drv_floors:
			fmt.Print("Arrived")
			fsm.FsmFloorArrival(floor)

		default:
			if timer.TimerTimedOut() && !obst { // Check for timeout only if no obstruction
				timer.TimerStop()
				fsm.FsmDoorTimeout()
			}
		}
	}
}

func main() {
	fmt.Println("Started!")
	hwelevio.Init(elevio.Addr, elevio.NFloors)

	if elevio.InputDevice.FloorSensor() == -1 {
		fsm.FsmInitBetweenFloors()
	}

	drv_buttons := make(chan hwelevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go hwelevio.PollButtons(drv_buttons)
	go hwelevio.PollFloorSensor(drv_floors)
	go hwelevio.PollObstructionSwitch(drv_obstr)
	go hwelevio.PollStopButton(drv_stop)

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
				fsm.FsmStop(stop)
			}
			fmt.Print("Done STPOPPPPP")

		case btnEvent := <-drv_buttons:
			fmt.Print("Button pressed")
			// Process button presses only if not stopped
			fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button)

		case floor := <-drv_floors:
			fmt.Print("Arrived")
			fsm.FsmFloorArrival(floor)

		default:
			if timer.TimerTimedOut() && !obst { // Check for timeout only if no obstruction
				timer.TimerStop()
				fsm.FsmDoorTimeout()
			}
		}
	}
	//var mainFuncObject processpair.MainFuncType = mainLogic
	//processpair.ProcessPairHandler(mainFuncObject)
}

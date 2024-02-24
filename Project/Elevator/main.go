package main

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
	"heislab/Elevator/elevio"
	"heislab/Elevator/fsm"
	"heislab/Elevator/processpair"
	"heislab/Elevator/timer"
	"time"
)

func mainLogic() {
	fsm.FsmResumeAtLatestCheckpoint()
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

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
			fsm.FsmMakeCheckpoint()
		}
	}
}

func main() {
	//hwelevio.Init(elevio.Addr, elevio.NFloors)
	//var mainFuncObject processpair.MainFuncType = fsm.FsmTestProcessPair

	fmt.Println("Started!")
	hwelevio.Init(elevio.Addr, elevio.NFloors)

	if elevio.InputDevice.FloorSensor() == -1 {
		fsm.FsmInitBetweenFloors()
		fsm.FsmMakeCheckpoint()
	}
	time.Sleep(1 * time.Second)
	var mainFuncObject processpair.MainFuncType = mainLogic
	processpair.ProcessPairHandler(mainFuncObject)

	// Block the main goroutine indefinitely
	done := make(chan struct{})
	<-done
}

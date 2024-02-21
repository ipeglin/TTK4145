package main

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
	"heislab/Elevator/elevio"
	"heislab/Elevator/fsm"
	"heislab/Elevator/timer"
)

func main() {
	//print?
	fmt.Println("Started!")
	//constants
	hwelevio.Init(elevio.Addr, elevio.NFloors)
	input := elevio.ElevioGetInputDevice()

	if input.FloorSensor() == -1 {
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
		//TODO
		case drv_obst := <-drv_obstr:
			obst = drv_obst

		//TODO
		case drv_stp := <-drv_stop:
			stop = drv_stp

		case btnEvent := <-drv_buttons:
			fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button)

		case floor := <-drv_floors:
			fsm.FsmFloorArrival(floor)

		default:
			if stop {
				fsm.FsmStop(stop)
			} else {
				fsm.FsmStop(false)
			}
			if obst {
				fsm.FsmObstruction()
			}
			if timer.TimerTimedOut() {
				timer.TimerStop()
				fsm.FsmDoorTimeout()
			}
		}
	}
}

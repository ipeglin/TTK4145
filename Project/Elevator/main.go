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

	for {
		select {
		case btnEvent := <-drv_buttons:
			fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button)

		case floor := <-drv_floors:
			fsm.FsmFloorArrival(floor)

		//TODO
		//case obstr:= <- drv_obstr.
		//	fsm.FsmObstruction()

		//TODO
		//case stop:= <- drv_stop:
		//	fsm.FsmStop()

		default:
			if timer.TimerTimedOut() {
				timer.TimerStop()
				fsm.FsmDoorTimeout()
			}
		}
	}
}

/*	for {
		{ //Request Button
			for f := 0; f < elevio.NFloors; f++ {
				for btn := hwelevio.BHallUp; btn < hwelevio.Last; btn++ {
					v := input.RequestButton(f, btn)
					if v && v != prev[f][btn] {
						fmt.Printf("Button has been requested")
						fsm.FsmRequestButtonPress(f, btn)
					}
					prev[f][btn] = v
				}
			}
		}

		{ //Floor sensor
			f := input.FloorSensor()
			fmt.Println("Floor: ", f)
			if f != -1 && f != prevFloor {
				fsm.FsmFloorArrival(f)
			}
			prevFloor = f
		}

		{ // Timer
			if timer.TimerTimedOut() {
				timer.TimerStop()
				fsm.FsmDoorTimeout()
			}
		}

		time.Sleep(time.Duration(elev.InputPollRateMsConfig) * time.Millisecond)
	}
}*/

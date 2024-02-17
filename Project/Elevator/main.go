/*package main

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
	"heislab/Elevator/elev"
	"heislab/Elevator/elevio"
	"heislab/Elevator/fsm"
	"heislab/Elevator/timer"
	"time"
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
	fmt.Println("Elevator is initialized")
	var prev [elevio.NFloors][elevio.NButtons]bool
	var prevFloor = -1
	for {
		{ //Request Button
			for f := 0; f < elevio.NFloors; f++ {
				for btn := elevio.BHallUp; btn < elevio.Last; btn++ {
					v := input.RequestButton(f, btn)
					if v && v != prev[f][btn] {
						fmt.Printf("Button has been requested")
						fsm.FsmRequestButtonPress(f, btn) //Dette er dårlig løsning MÅ FIKSE
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

package main

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
	"heislab/Elevator/elev"
	"heislab/Elevator/elevio"
	"heislab/Elevator/fsm"
	"heislab/Elevator/timer"
	"time"
)

func main() {
	fmt.Println("Started!")
	//hwelevio.Init(elevio.Addr, elevio.NFloors)
	input := elevio.ElevioGetInputDevice()

	if input.FloorSensor() == -1 {
		fsm.FsmInitBetweenFloors()
	}
	fmt.Println("Elevator is initialized")

	// Create channels for button events and floor sensor events
	buttonEvents := make(chan hwelevio.ButtonEvent)
	floorEvents := make(chan int)

	// Start polling goroutines
	go pollButtons(input, buttonEvents)
	go pollFloorSensor(input, floorEvents)

	for {
		select {
		case btnEvent := <-buttonEvents:
			fmt.Println("Button has been requested:", btnEvent)
			fsm.FsmRequestButtonPress(btnEvent.Floor, btnEvent.Button)

		case floor := <-floorEvents:
			fmt.Println("Floor sensor detected arrival at:", floor)
			fsm.FsmFloorArrival(floor)

		default:
			// Timer check
			if timer.TimerTimedOut() {
				timer.TimerStop()
				fsm.FsmDoorTimeout()
			}
			time.Sleep(time.Duration(elev.InputPollRateMsConfig) * time.Millisecond)
		}
	}
}

func pollButtons(input elevio.ElevInputDevice, buttonEvents chan<- hwelevio.ButtonEvent) {
	prev := make([][elevio.NButtons]bool, elevio.NFloors)
	for {
		for f := 0; f < elevio.NFloors; f++ {
			for btn := hwelevio.BHallUp; btn < hwelevio.Last; btn++ {
				v := input.RequestButton(f, btn)
				if v && !prev[f][btn] {
					buttonEvents <- hwelevio.ButtonEvent{Floor: f, Button: btn}
				}
				prev[f][btn] = v
			}
		}
		time.Sleep(time.Duration(elev.InputPollRateMsConfig) * time.Millisecond)
	}
}

func pollFloorSensor(input elevio.ElevInputDevice, floorEvents chan<- int) {
	prevFloor := -1
	for {
		f := input.FloorSensor()
		fmt.Println("Floor: ", f)
		if f != -1 && f != prevFloor {
			floorEvents <- f
		}
		prevFloor = f
		time.Sleep(time.Duration(elev.InputPollRateMsConfig) * time.Millisecond)
	}
}

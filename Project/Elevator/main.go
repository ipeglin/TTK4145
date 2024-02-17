package main

import (
	"fmt"
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
				for btn := 0; btn < elevio.NButtons; btn++ {
					v := input.RequestButton(f, btn)
					if v && v != prev[f][btn] {
						fmt.Printf("Button has been requested")
						fsm.FsmRequestButtonPress(f, elevio.Button(btn)) //Dette er dårlig løsning MÅ FIKSE
					}
					prev[f][btn] = v
				}
			}
		}

		{ //Floor sensor
			f := input.FloorSensor()
			fmt.Print(f)
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
}

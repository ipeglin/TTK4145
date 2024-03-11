package elevio

import (
	"elevator/driver/hwelevio"
	"elevator/timer"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

var InputDevice ElevInputDevice

func init() {
	hwelevio.Init(Addr, NFloors)
	InputDevice = ElevioGetInputDevice()
}

// Liker ikke dene løsningen
func castElevDirToMotorDirection(d ElevDir) hwelevio.HWMotorDirection {
	switch d {
	case DirDown:
		return hwelevio.MD_Down
	case DirUp:
		return hwelevio.MD_Up
	case DirStop:
		return hwelevio.MD_Stop
	default:
		fmt.Printf("Noe har gått feil i CastMotorDirection")
		return hwelevio.MD_Down //HELT FEIL MÅ FIKSES
	}
}

func castButtonToHWButtonType(btn Button) hwelevio.HWButtonType {
	switch btn {
	case BHallUp:
		return hwelevio.BHallUp
	case BHallDown:
		return hwelevio.BHallDown
	case BCab:
		return hwelevio.BCab
	default:
		logrus.Error("Something went wrong!")
		return hwelevio.BHallUp //Hvordan løse dette?
	}
}

// New name?
func RequestButton(f int, btn Button) bool {
	// Implementation using the actual hardware library.
	return hwelevio.GetButton(castButtonToHWButtonType(btn), f)
}

func RequestButtonLight(f int, btn Button, v bool) {
	// Implementation using the actual hardware library.
	hwelevio.SetButtonLamp(castButtonToHWButtonType(btn), f, v)
}

func RequestFloorIndicator(f int) {
	hwelevio.SetFloorIndicator(f)
}

func RequestDoorOpenLamp(v bool) {
	hwelevio.SetDoorOpenLamp(v)
}

func RequestStopLamp(v bool) {
	hwelevio.SetStopLamp(v)
}

func RequestObstruction() bool {
	return hwelevio.GetObstruction()
}

func RequestFloor() int {
	return hwelevio.GetFloor()
}

func RequestStop() bool {
	return hwelevio.GetStop()
}

func MotorDirection(d ElevDir) {
	// Implementation using the actual hardware library.
	hwelevio.SetMotorDirection(castElevDirToMotorDirection(d))
}

func ElevioGetInputDevice() ElevInputDevice {
	return ElevInputDevice{
		FloorSensor:   RequestFloor,
		RequestButton: RequestButton,
		StopButton:    RequestStop,
		Obstruction:   RequestObstruction,
	}
}

func ElevioGetOutputDevice() ElevOutputDevice {
	return ElevOutputDevice{
		FloorIndicator:     RequestFloorIndicator,
		RequestButtonLight: RequestButtonLight,
		DoorLight:          RequestDoorOpenLamp,
		StopButtonLight:    RequestStopLamp,
		MotorDirection:     MotorDirection,
	}
}

func ElevDirToString(d ElevDir) string {
	switch d {
	case DirDown:
		return "down"
	case DirStop:
		return "stop"
	case DirUp:
		return "up"
	default:
		return "DirUnknown"
	}
}

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, NFloors)
	for {
		time.Sleep(hwelevio.PollRate)
		for f := 0; f < NFloors; f++ {
			for b := Button(0); b < 3; b++ {
				v := hwelevio.GetButton(castButtonToHWButtonType(b), f)
				if v != prev[f][b] && v {
					receiver <- ButtonEvent{f, Button(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(hwelevio.PollRate)
		v := hwelevio.GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(hwelevio.PollRate)
		v := hwelevio.GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(hwelevio.PollRate)
		v := hwelevio.GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func MontitorMotorActivity(receiver chan<- bool, duration float64) {
	timerActive := true
	timerEndTimer := timer.GetCurrentTimeAsFloat() + duration
	for {
		time.Sleep(hwelevio.PollRate)
		v := RequestFloor()
		if v != -1 {
			timerEndTimer = timer.GetCurrentTimeAsFloat() + duration
			if !timerActive {
				timerActive = true
				receiver <- true
			}
		} else {
			if timer.GetCurrentTimeAsFloat() > timerEndTimer {
				if timerActive {
					timerActive = false
					receiver <- false
				}
			}
		}
	}
}

func ButtonToString(b Button) string {
	switch b {
	case BHallUp:
		return "BHallUp"
	case BHallDown:
		return "BHallDown"
	case BCab:
		return "BCab"
	default:
		return "Button Unknown"
	}
}

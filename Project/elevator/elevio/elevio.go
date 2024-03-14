package elevio

import (
	"elevator/elevio/driver/hwelevio"
	"time"

	"github.com/sirupsen/logrus"
)

var InputDevice ElevInputDevice

func init() {
	hwelevio.Init(Addr)
	InputDevice = ElevioGetInputDevice()
}

func castElevDirToMotorDirection(d ElevDir) hwelevio.HWMotorDirection {
	switch d {
	case DirDown:
		return hwelevio.MDDown
	case DirUp:
		return hwelevio.MDUp
	case DirStop:
		return hwelevio.MDStop
	default:
		logrus.Error("Something went wrong!")
		return hwelevio.MDDown
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
		return hwelevio.BHallUp
	}
}

func RequestButton(f int, btn Button) bool {
	return hwelevio.GetButton(castButtonToHWButtonType(btn), f)
}

func RequestButtonLight(f int, btn Button, v bool) {
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

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, NFloors)
	for {
		time.Sleep(PollRateMS * time.Millisecond)
		for f := 0; f < NFloors; f++ {
			for b := BHallUp; b <= BCab; b++ {
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
		time.Sleep(PollRateMS * time.Millisecond)
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
		time.Sleep(PollRateMS * time.Millisecond)
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
		time.Sleep(PollRateMS * time.Millisecond)
		v := hwelevio.GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
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

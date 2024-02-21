package elevio

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
)

func init() {
	hwelevio.Init(Addr, NFloors)
}

// Liker ikke dene løsningen
func castElevDirToMotorDirection(d ElevDir) hwelevio.MotorDirection {
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

// New name?
func RequestButton(f int, btn hwelevio.Button) bool {
	// Implementation using the actual hardware library.
	return hwelevio.GetButton(btn, f)
}

func RequestButtonLight(f int, btn hwelevio.Button, v bool) {
	// Implementation using the actual hardware library.
	hwelevio.SetButtonLamp(btn, f, v)
}

func MotorDirection(d ElevDir) {
	// Implementation using the actual hardware library.
	fmt.Println("Motordirection to be sat: ", ElevDirToString(d))
	hwelevio.SetMotorDirection(castElevDirToMotorDirection(d))
}

func ElevioGetInputDevice() ElevInputDevice {
	return ElevInputDevice{
		FloorSensor:   hwelevio.GetFloor,
		RequestButton: RequestButton,
		StopButton:    hwelevio.GetStop,
		Obstruction:   hwelevio.GetObstruction,
	}
}

func ElevioGetOutputDevice() ElevOutputDevice {
	return ElevOutputDevice{
		FloorIndicator:     hwelevio.SetFloorIndicator,
		RequestButtonLight: RequestButtonLight,
		DoorLight:          hwelevio.SetDoorOpenLamp,
		StopButtonLight:    hwelevio.SetStopLamp,
		MotorDirection:     MotorDirection,
	}
}

func ElevDirToString(d ElevDir) string {
	switch d {
	case DirDown:
		return "DirDown"
	case DirStop:
		return "DirStop"
	case DirUp:
		return "DirUp"
	default:
		return "DirUnknown"
	}
}

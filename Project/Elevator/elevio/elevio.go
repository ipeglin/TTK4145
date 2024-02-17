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

// Trenger nok ne bedre løsning her
func castIntToBtnType(btn_int int) hwelevio.ButtonType {
	switch btn_int {
	case 0:
		return hwelevio.BT_HallDown
	case 1:
		return hwelevio.BT_HallUp
	case 2:
		return hwelevio.BT_Cab
	default:
		//IDK
		fmt.Printf("Noe har gått feil i CastBtnType")
		return hwelevio.BT_Cab //Dette er ikke pent
	}
}

// New name?
func RequestButton(f int, btn_int int) bool {
	// Implementation using the actual hardware library.
	return hwelevio.GetButton(castIntToBtnType(btn_int), f)
}

func RequestButtonLight(f int, btn_int int, v bool) {
	// Implementation using the actual hardware library.
	hwelevio.SetButtonLamp(castIntToBtnType(btn_int), f, v)
}

func MotorDirection(d ElevDir) {
	// Implementation using the actual hardware library.
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

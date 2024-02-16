package elevio

import (
	"heislab/Elevator/driver/hwelevio"
)

func init() {
	hwelevio.Init(Addr, NFloors)
}

// New name?
func RequestButton(f int, b Button) bool {
	// Implementation using the actual hardware library.
	return hwelevio.GetButton(b, f)
}

func RequestButtonLight(f int, b Button, v bool) {
	// Implementation using the actual hardware library.
	hwelevio.SetButtonLamp(b, f, v)
}

func MotorDirection(d ElevDir) {
	// Implementation using the actual hardware library.
	hwelevio.SetMotorDirection(d)
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

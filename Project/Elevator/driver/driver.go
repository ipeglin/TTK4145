package driver

import (
	"fmt"
	"heislab/Elevator/driver/hwelevio"
)

func Driver() {

	numFloors := 4

	hwelevio.Init("localhost:15657", numFloors)

	var d hwelevio.MotorDirection = hwelevio.MD_Up
	//hwelevio.SetMotorDirection(d)

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
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			hwelevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = hwelevio.MD_Down
			} else if a == 0 {
				d = hwelevio.MD_Up
			}
			hwelevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				hwelevio.SetMotorDirection(hwelevio.MD_Stop)
			} else {
				hwelevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := hwelevio.Button(0); b < 3; b++ {
					hwelevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}

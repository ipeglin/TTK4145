package elev

import (
	"fmt"
	"elevator/driver/hwelevio"
	"elevator/elevio"
)

func ElevatorInit() Elevator {
	return Elevator{
		CurrentFloor:     -1,
		CurrentBehaviour: EBIdle,
		Config: ElevatorConfig{
			ClearRequestVariant: ClearRequestVariantConfig,
			DoorOpenDurationS:   DoorOpenDurationSConfig,
		},
	}
}

func ElevatorPrint(e Elevator) {
	fmt.Println("\n  +--------------------+")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12s|\n"+
			"  |behav = %-12s|\n",
		e.CurrentFloor,
		elevio.ElevDirToString(e.Dirn), // Assuming this function exists
		ebToString(e.CurrentBehaviour), // You'll need to implement or assume this function
	)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := elevio.NFloors - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := hwelevio.BHallUp; btn <= hwelevio.BCab; btn++ {
			if (f == elevio.NFloors-1 && btn == hwelevio.BHallUp) ||
				(f == 0 && btn == hwelevio.BHallDown) {
				fmt.Print("|     ")
			} else {
				if e.Requests[f][btn] {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

// Assuming ebToString function exists or is defined similar to:
func ebToString(behaviour ElevatorBehaviour) string {
	switch behaviour {
	case EBIdle:
		return "idle"
	case EBDoorOpen:
		return "doorOpen"
	case EBMoving:
		return "moving"
	default:
		return "Unknown"
	}
}

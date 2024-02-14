package main

func elevator_init() Elevator {
	return (Elevator) {
		CurrentFloor = -1
		CurrentBehaviour = EBIdle
		Config = {
			ClearRequestVariant = ClearRequestVariantConfig
			DoorOpenDurationS = DoorOpenDurationSConfig
		}
	}
}

type Elevator struct {
	CurrentFloor int
	Requests [N_FLOORS][N_BUTTONS]int
	CurrentBehaviour ElevatorBehaviour

	Config struct {
        ClearRequestVariant ClearRequestVariant
        DoorOpenDurationS   float64
    }
}
package elevator

func ElevatorInit() Elevator {
	return Elevator{
		CurrentFloor:     -1,
		CurrentBehaviour: EBIdle,
		Config: struct {
			ClearRequestVariant ClearRequestVariant
			DoorOpenDurationS   float64
		}{
			ClearRequestVariant: ClearRequestVariantConfig,
			DoorOpenDurationS:   DoorOpenDurationSConfig,
		},
	}
}

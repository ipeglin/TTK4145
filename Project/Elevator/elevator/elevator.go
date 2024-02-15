package elevator

func elevator_init() Elevator {
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

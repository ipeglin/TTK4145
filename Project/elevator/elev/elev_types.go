package elev

import (
	"elevator/elevio"
)

type ElevatorBehaviour int

const (
	EBIdle ElevatorBehaviour = iota
	EBDoorOpen
	EBMoving
)

type ClearRequestVarient int

const (
	CRVAll ClearRequestVarient = iota
	CRVInDirn
)

type ElevatorConfig struct {
	ClearRequestVariant ClearRequestVarient
	DoorOpenDurationS   float64
}

type Elevator struct {
	CurrentFloor     int
	Dirn             elevio.ElevDir
	Requests         [elevio.NFloors][elevio.NButtons]bool
	CurrentBehaviour ElevatorBehaviour
	Config           ElevatorConfig
}

package elev

import (
	"elevator/elevio"
)

// Define ElevatorBehaviour as a custom type
type ElevatorBehaviour int

// Declare states using iota
const (
	EBIdle ElevatorBehaviour = iota
	EBDoorOpen
	EBMoving
)

// Define ClearRequestVarient as a custom type
type ClearRequestVarient int

// Declare states using iota
const (
	CRVAll ClearRequestVarient = iota
	CRVInDirn
)

type ElevatorConfig struct {
	ClearRequestVariant ClearRequestVarient
	DoorOpenDurationS   float64
}

// Elevator struct to represent the state machine
type Elevator struct {
	CurrentFloor     int
	Dirn             elevio.ElevDir
	Requests         [elevio.NFloors][elevio.NButtons]bool
	CurrentBehaviour ElevatorBehaviour
	Config           ElevatorConfig
}

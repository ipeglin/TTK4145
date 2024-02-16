package elevator

import (
	"heislab/Elevator/elevio"
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

// Elevator struct to represent the state machine
type Elevator struct {
	CurrentFloor     int
	Requests         [elevio.NFloors][elevio.NButtons]int
	CurrentBehaviour ElevatorBehaviour

	Config struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDurationS   float64
	}
}

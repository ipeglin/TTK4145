package main

import (

)

// Define ElevatorBehaviour as a custom type
type ElevatorBehaviour int

// Declare states using iota
const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVarient int

const (
	CRV_All ClearRequestVarient = iota
	CRV_InDirn
)

// Elevator struct to represent the state machine
type Elevator struct {
	CurrentFloor     int
	RequestedFloor   int
	CurrentBehaviour ElevatorBehaviour
}

func main {

}
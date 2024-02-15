package requests

import (
	"heislab/Elevator/elevator"
	"heislab/Elevator/elevator_io"
)

type DirnBehaviourPair struct {
	dirn      elevator_io.Dirn
	behaviour elevator.ElevatorBehaviour
}

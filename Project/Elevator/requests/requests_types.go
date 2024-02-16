package requests

import (
	"heislab/Elevator/elevator"
	"heislab/Elevator/elevio"
)

type DirnBehaviourPair struct {
	dirn      elevio.ElevDir
	behaviour elevator.ElevatorBehaviour
}

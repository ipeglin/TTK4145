package requests

import (
	"heislab/Elevator/elev"
	"heislab/Elevator/elevio"
)

type DirnBehaviourPair struct {
	Dirn      elevio.ElevDir
	Behaviour elev.ElevatorBehaviour
}

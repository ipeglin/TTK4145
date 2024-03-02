package requests

import (
	"elevator/elev"
	"elevator/elevio"
)

type DirnBehaviourPair struct {
	Dirn      elevio.ElevDir
	Behaviour elev.ElevatorBehaviour
}

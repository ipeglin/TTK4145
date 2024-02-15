package elevator_io

const (
	NFloors  = 4
	NButtons = 3
)

type Dirn int

const (
	DirDown Dirn = iota - 1
	DirStop
	DirUp
)

type Button int

const (
	BHallUp Button = iota
	BHallDown
	BCab
)

type ElevInputDevice struct {
}

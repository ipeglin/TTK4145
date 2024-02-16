package elevio

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

// ElevInputDevice defines the interface for elevator input devices.
type ElevInputDevice struct {
	FloorSensor   func() int
	RequestButton func(floor int, b Button) int
	StopButton    func() int
	Obstruction   func() int
}

// ElevOutputDevice defines the interface for elevator output devices.
type ElevOutputDevice struct {
	FloorIndicator     func(floor int)
	RequestButtonLight func(floor int, b Button, on int)
	DoorLight          func(on int)
	StopButtonLight    func(on int)
	MotorDirection     func(d Dirn)
}

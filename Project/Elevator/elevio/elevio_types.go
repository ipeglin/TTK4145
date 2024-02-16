package elevio

const (
	NFloors  int = 4
	NButtons int = 3
)

const Addr string = "localhost:15657"

type ElevDir int

const (
	DirDown ElevDir = iota - 1
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
	RequestButton func(floor int, b Button) bool
	StopButton    func() bool
	Obstruction   func() bool
}

// ElevOutputDevice defines the interface for elevator output devices.
type ElevOutputDevice struct {
	FloorIndicator     func(floor int)
	RequestButtonLight func(floor int, b Button, on int)
	DoorLight          func(on int)
	StopButtonLight    func(on int)
	MotorDirection     func(d ElevDir)
}

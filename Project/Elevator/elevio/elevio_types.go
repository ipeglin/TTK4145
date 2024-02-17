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
	RequestButton func(f int, btn_int int) bool
	StopButton    func() bool
	Obstruction   func() bool
}

// ElevOutputDevice defines the interface for elevator output devices.
type ElevOutputDevice struct {
	FloorIndicator     func(f int)
	RequestButtonLight func(f int, btn_int int, v bool)
	DoorLight          func(v bool)
	StopButtonLight    func(v bool)
	MotorDirection     func(d ElevDir)
}

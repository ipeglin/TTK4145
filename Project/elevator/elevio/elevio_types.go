package elevio

const (
	NFloors    int = 4
	NButtons   int = 3
	PollRateMS     = 20
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

type ButtonEvent struct {
	Floor  int
	Button Button
}

type inputDevice struct {
	FloorSensor   func() int
	RequestButton func(f int, btn Button) bool
	StopButton    func() bool
	Obstruction   func() bool
}

type outputDevice struct {
	FloorIndicator     func(f int)
	RequestButton      func(f int, btn Button) bool
	RequestButtonLight func(f int, btn Button, v bool)
	DoorLight          func(v bool)
	StopButtonLight    func(v bool)
	MotorDirection     func(d ElevDir)
}

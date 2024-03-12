package hwelevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const PollRate = 20 * time.Millisecond

// TODO: refactor these to be in line with camelCase
var initialized bool = false
var nFloors int = 4
var mtx sync.Mutex
var conn net.Conn

type HWMotorDirection int

// TODO: refactor these to be in line with camelCase
const (
	MDDown HWMotorDirection = iota - 1
	MDStop
	MDUp
)

type HWButtonType int

const (
	BHallUp HWButtonType = iota
	BHallDown
	BCab
)

func Init(addr string, numFloors int) {
	if initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	nFloors = numFloors
	mtx = sync.Mutex{}
	var err error
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	initialized = true
}

func SetMotorDirection(dir HWMotorDirection) {
	//fmt.Println("Setting motordirection")
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button HWButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func GetButton(button HWButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	mtx.Lock()
	defer mtx.Unlock()

	_, err := conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	mtx.Lock()
	defer mtx.Unlock()

	_, err := conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

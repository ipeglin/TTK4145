// Use `go run foo.go` to run your program

package main

import (
	"fmt"
	"runtime"
)

var i = 0

func incrementing(op chan string, quit chan bool) {
	for j := 0; j < 1_000_5; j++ {
		op <- "increment"
	}
	quit <- true
}

func decrementing(op chan string, quit chan bool) {
	for j := 0; j < 1_000_0; j++ {
		op <- "decrement"
	}
	quit <- true
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(3)
	op := make(chan string)
	quit := make(chan bool)
	// TODO: Spawn both functions as goroutines

	go incrementing(op, quit)
	go decrementing(op, quit)

	go func() {
		for {
			select {
			case msg := <-op:
				if msg == "increment" {
					i++

				} else if msg == "decrement" {
					i--
				}
			default:
				//hei
			}
		}
	}()
	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.
	<-quit
	<-quit
	fmt.Println("The magic number is:", i)
}

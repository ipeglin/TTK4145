package main

import (
	// "elevator"
	"flag"
	"fmt"
	"network"
	"network/nodes"
	"os"
	"time"
	"watchdog"

	"github.com/sirupsen/logrus"
)

// fetching process flags
func getFlags() (int, error) {
	var watch *int
	watch = flag.Int("watch", 0, "watch process with given id")
	flag.Parse()

	// require id
	if *watch == 0 {
		return 0, nil
	}

	return *watch, nil
}

func main() {
	watch, err := getFlags()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	
	processId := os.Getpid()
	if watch != 0 {
		fmt.Println(processId) // pass PID to watchdog with pipeline
	}
	logrus.Info("Initialising node PID: ", processId)

	// BUG! Backup process is the one that proceedes, not the main process
	done := make(chan bool)
	go watchdog.Init(watch, done)
	<-done

	// TODO: Launch new process watching current process in case of crash

	nodeOverviewChannel := make(chan nodes.NetworkNodeRegistry)
	messageReceiveChannel := make(chan network.Message)
	messageTransmitterChannel := make(chan network.Message)
	onlineStatusChannel := make(chan bool)

	// must be commented outside of lab
	// go elevator.Init()

	go network.Init(nodeOverviewChannel, messageTransmitterChannel, messageReceiveChannel, onlineStatusChannel)
	go func(){
		for {
			messageTransmitterChannel <- network.Message{Payload: fmt.Sprintf("Hello World from process %v", processId), MessageId: 0}
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case reg := <-nodeOverviewChannel:
			logrus.Info("Known nodes:", reg.Nodes)
		case msg := <-messageReceiveChannel:
			fmt.Printf("Must do something with the message:%v\n", msg.Payload)
		case online:= <-onlineStatusChannel:
			logrus.Warn("Updated online status:", online)
		}
	}
}

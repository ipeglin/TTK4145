package main

import (
	"fmt"
	"network"
	"network/nodes"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Node initialised with PID:", os.Getpid())

	// TODO: Launch new process watching current process in case of crash

	nodeOverviewChannel := make(chan nodes.NetworkNodeRegistry)
	messageReceiveChannel := make(chan network.Message)
	messageTransmitterChannel := make(chan network.Message)
	onlineStatusChannel := make(chan bool)

	go network.Init(nodeOverviewChannel, messageTransmitterChannel, messageReceiveChannel, onlineStatusChannel)
	go func(){
		for {
			messageTransmitterChannel <- network.Message{Payload: fmt.Sprintf("Hello World from %s", "This unit"), MessageId: 0}
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

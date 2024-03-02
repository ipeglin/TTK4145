package main

import (
	"flag"
	"fmt"
	"network"
	"network/nodes"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// fetching process flags
func getFlags() (int, int, error) {
	var id, numNodes *int
	id = flag.Int("id", 0, "ID of the node")
	numNodes = flag.Int("numNodes", 3, "total number of the nodes")
	flag.Parse()

	// require id
	if *id == 0 {
		return 0, 0, fmt.Errorf("ERROR: Node ID is required")
	} else if *id > *numNodes {
		return 0, 0, fmt.Errorf("ERROR: Node ID cannot be greater than number of nodes")
	}

	return *id, *numNodes, nil
}

func main() {
	id, numNodes, err := getFlags()
	if err != nil {
		logrus.Fatal(err)
		return
	}

	logrus.Debug(fmt.Sprintf("Node initialised with ID=%d, #Nodes=%d, PID=%d", id, numNodes, os.Getpid()))

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
			logrus.Warn("Updated nodes:", reg.Nodes)
		case msg := <-messageReceiveChannel:
			logrus.Warn("Broadcast received:", msg)
		case online:= <-onlineStatusChannel:
			logrus.Warn("Updated online status: %v\n", online)
		}
	}
}

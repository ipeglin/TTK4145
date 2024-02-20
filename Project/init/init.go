package main

import (
	"flag"
	"fmt"
	"network"
	"network/nodes"
	"os"
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
		fmt.Println(err)
		return
	}

	fmt.Printf("Node initialised with:\n")
	fmt.Printf("  ID:          %d\n", id)
	fmt.Printf("  #Nodes:      %d\n", numNodes)
	fmt.Printf("  PID:         %d\n", os.Getpid())

	// TODO: Launch new process watching current process in case of crash

	nodeOverviewChannel := make(chan nodes.NetworkNodeRegistry)
	messageReceiveChannel := make(chan network.Message)
	messageTransmitterChannel := make(chan network.Message)

	go network.Init(nodeOverviewChannel, messageTransmitterChannel, messageReceiveChannel)

	for {
		select {
		case reg := <-nodeOverviewChannel:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", reg.Nodes)
			fmt.Printf("  New:      %q\n", reg.New)
			fmt.Printf("  Lost:     %q\n", reg.Lost)

		case msg := <-messageReceiveChannel:
			fmt.Printf("Network module says:%v\n", msg)
		}
	}
}

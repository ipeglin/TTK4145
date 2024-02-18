package main

import (
	"flag"
	"fmt"
	"network/local"
	"network/nodes"
	"os"
)

type Message struct {
	Content    string
	Iterations int
}

func main() {
	// fetching process flags
	var id, numNodes, basePort int
	flag.IntVar(&id, "nodeID", 0, "ID of the node")
	flag.Parse()

	// id flag required
	if id == 0 {
		fmt.Println("ERROR: Node ID is required")
		return
	} else if id > numNodes {
		fmt.Println("WARNING: Node ID should not be greater than total number of nodes")
	}

	flag.IntVar(&numNodes, "numNodes", 3, "total number of the nodes")
	flag.Parse()

	flag.IntVar(&basePort, "basePort", 1337, "Base port for all nodes")
	flag.Parse()

	// setting up the ports
	const portsPerHost int = 2
	var lifelinePort int

	// fetching host IP and PORT
	nodeIP, err := local.GetIP()
	if err != nil {
		fmt.Println("ERROR: Unable to get the IP address")
		nodeIP = "Disconnected"
	} else {
		lifelinePort = basePort + (id-1)*portsPerHost + 1 // first port number of the node
	}

	// set node unique ID
	nodeUid := fmt.Sprintf("peer-%s-%d", nodeIP, os.Getpid())

	// channel for network node updates
	nodeRegistryChannel := make(chan nodes.NetworkNodeRegistry)
	TransmissionEnableChannel := make(chan bool)

	go nodes.Client(lifelinePort, nodeUid, TransmissionEnableChannel)
	go nodes.Server(lifelinePort, nodeRegistryChannel)

	// // broadcast message sender and receiver
	// messageSender := make(chan Message)
	// messageReceiver := make(chan Message)

	// var broadcastPort int = lifelinePort + 1
	// go broadcast.Client(broadcastPort, messageSender)
	// go broadcast.Server(broadcastPort, messageReceiver)

	// // Taken from https://github.com/TTK4145/Network-go/blob/master/main.go
	// // The example message. We just send one of these every second.
	// go func() {
	// 	message := Message{"Hello from " + nodeUid, 0}
	// 	for {
	// 		message.Iterations++
	// 		messageSender <- message
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	fmt.Println("Network module starting...")
	for {
		select {
		case reg := <-nodeRegistryChannel:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", reg.Nodes)
			fmt.Printf("  New:      %q\n", reg.New)
			fmt.Printf("  Lost:     %q\n", reg.Lost)

			// case msg := <-messageReceiver:
			// 	fmt.Printf("Received: %#v\n", msg)
		}
	}
}

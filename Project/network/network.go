package network

import (
	"fmt"
	"network/broadcast"
	"network/local"
	"network/nodes"
)

const basePort int = 1337
const lifelinePort int = basePort + 1
const transmissionPort int = basePort + 2
const receiverPort int = basePort + 3

type Message struct {
	Content    string
	Iterations int
}

func Init(nodesChannel chan<- nodes.NetworkNodeRegistry, messageChannel <-chan Message, responseChannel chan<- Message) {
	fmt.Println("Initialising Network Module...")

	// fetching host IP and PORT
	nodeIP, err := local.GetIP()
	if err != nil {
		fmt.Println("ERROR: Unable to get the IP address")
		nodeIP = "Disconnected"
	}

	// set node unique ID
	nodeUid := fmt.Sprintf("peer-%s", nodeIP)

	fmt.Printf("Module initialised with:\n")
	fmt.Printf("  IPv4:     %v\n", nodeIP)
	fmt.Printf("  PORT:     %d\n", basePort)
	fmt.Printf("  UID:      %s\n", nodeUid)

	// channel for network node updates
	nodeRegistryChannel := make(chan nodes.NetworkNodeRegistry)
	TransmissionEnableChannel := make(chan bool)

	go nodes.Client(lifelinePort, nodeUid, TransmissionEnableChannel)
	go nodes.Server(lifelinePort, nodeRegistryChannel)

	broadcastTransmissionChannel := make(chan Message)
	broadcastReceiverChannel := make(chan Message)

	go broadcast.Client(transmissionPort, broadcastTransmissionChannel)
	go broadcast.Server(receiverPort, broadcastReceiverChannel)

	fmt.Println("Network module starting...")
	for {
		select {
		case reg := <-nodeRegistryChannel:
			nodesChannel <- reg

		case msg := <-broadcastReceiverChannel:
			responseChannel <- msg

		case msg := <-messageChannel:
			fmt.Println("Network module intercepted message:", msg.Content)
			broadcastTransmissionChannel <- msg
		}
	}
}

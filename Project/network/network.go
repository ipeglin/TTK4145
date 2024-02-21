package network

import (
	"fmt"
	"network/broadcast"
	"network/local"
	"network/nodes"
	"os"

	"github.com/sirupsen/logrus"
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
	logrus.Trace("Initialising Network Module...")

	// fetching host IP and PORT
	nodeIP, err := local.GetIP()
	if err != nil {
		logrus.Warn("ERROR: Unable to get the IP address")
		nodeIP = "Disconnected"
	}

	// set node unique ID
	nodeUid := fmt.Sprintf("peer-%s-%d", nodeIP, os.Getpid())
	logrus.Info(fmt.Sprintf("Network module initialised with UID=%s on PORT=%d", nodeUid, basePort))

	// channel for network node updates
	nodeRegistryChannel := make(chan nodes.NetworkNodeRegistry)
	TransmissionEnableChannel := make(chan bool)

	go nodes.Client(lifelinePort, nodeUid, TransmissionEnableChannel)
	go nodes.Server(lifelinePort, nodeRegistryChannel)

	broadcastTransmissionChannel := make(chan Message)
	broadcastReceiverChannel := make(chan Message)

	go broadcast.Client(transmissionPort, broadcastTransmissionChannel)
	go broadcast.Server(receiverPort, broadcastReceiverChannel)

	for {
		select {
		case reg := <-nodeRegistryChannel:
			logrus.Info("Node registry updated")
			nodesChannel <- reg

		case msg := <-broadcastReceiverChannel:
			logrus.Info("Broadcast received")
			responseChannel <- msg

		case msg := <-messageChannel:
			logrus.Debug("Network module intercepted message:", msg.Content)
			broadcastTransmissionChannel <- msg
		}
	}
}

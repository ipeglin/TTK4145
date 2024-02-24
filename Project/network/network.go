package network

import (
	"fmt"
	"network/broadcast"
	"network/local"
	"network/nodes"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const basePort int = 1337
const lifelinePort int = basePort + 1
const transmissionPort int = basePort + 2
const receiverPort int = basePort + 3

type Message struct {
	Content    string
	Iterations int
}

func Init(nodesChannel chan<- nodes.NetworkNodeRegistry, messageChannel <-chan Message, responseChannel chan<- Message, onlineStatusChannel chan<- bool) {
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
			logrus.Info(fmt.Sprintf("Node registry update:\n  Nodes:    %q\n  New:      %q\n  Lost:     %q", reg.Nodes, reg.New, reg.Lost))

			// pass node online status to the main process
			if slices.Contains(reg.Lost,  nodeUid) {
				logrus.Warn("Node lost connection: ", nodeUid)
				onlineStatusChannel <- false
			} else if reg.New == nodeUid {
				logrus.Info("Node connected: ", nodeUid)
				onlineStatusChannel <- true
			} 

			nodesChannel <- reg

		case msg := <-broadcastReceiverChannel:
			logrus.Debug("Broadcast received")
			responseChannel <- msg

		case msg := <-messageChannel:
			logrus.Debug("Network intercepted message")
			broadcastTransmissionChannel <- msg
		}
	}
}

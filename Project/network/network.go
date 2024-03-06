package network

import (
	"fmt"
	"network/broadcast"
	"network/checksum"
	"network/local"
	"network/nodes"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const basePort int = 1337
const lifelinePort int = basePort + 1
const messagePort int = basePort + 2

type Message struct {
	MessageId int
	SenderId  string // IPv4
	Payload   interface{}
	Checksum  string
}

func Init(nodesChannel chan<- nodes.NetworkNodeRegistry, messageChannel <-chan Message, responseChannel chan<- Message, onlineStatusChannel chan<- bool, ipChannel chan<- string) {
	// fetching host IP and PORT
	nodeIP, err := local.GetIP()
	if err != nil {
		logrus.Warn("ERROR: Unable to get the IP address")
	}

	ipChannel <- nodeIP

	// set node unique ID
	nodeUid := fmt.Sprintf("peer-%s-%d", nodeIP, os.Getpid())
	logrus.Debug(fmt.Sprintf("Network module initialised with UID=%s on PORT=%d", nodeUid, basePort))

	// channel for network node updates
	nodeRegistryChannel := make(chan nodes.NetworkNodeRegistry)
	TransmissionEnableChannel := make(chan bool)

	go nodes.Sender(lifelinePort, nodeUid, TransmissionEnableChannel)
	go nodes.Receiver(lifelinePort, nodeRegistryChannel)

	broadcastTransmissionChannel := make(chan Message)
	broadcastReceiverChannel := make(chan Message)

	go broadcast.Sender(messagePort, broadcastTransmissionChannel)
	go broadcast.Receiver(nodeIP, messagePort, broadcastReceiverChannel)

	for {
		select {
		case reg := <-nodeRegistryChannel:
			logrus.Debug(fmt.Sprintf("Node registry update:\n  Nodes:    %q\n  New:      %q\n  Lost:     %q", reg.Nodes, reg.New, reg.Lost))

			// pass node online status to the main process
			if slices.Contains(reg.Lost, nodeUid) {
				logrus.Warn("Node lost connection:", nodeUid)
				onlineStatusChannel <- false
			} else if reg.New == nodeUid {
				logrus.Warn("Node connected:", nodeUid)
				onlineStatusChannel <- true
			}

			nodesChannel <- reg

		case msg := <-broadcastReceiverChannel:
			logrus.Debug("Broadcast received from network")
			/*checksum, err := checksum.GenerateJSONChecksum(msg.Payload)
			if err != nil {
				logrus.Error("Checksum generation failed:", err)
				continue
			}

			if msg.Checksum != checksum {
				logrus.Error("Checksum mismatch, payload corrupted. Abort forwarding")
				continue
			}*/

			responseChannel <- msg

		case msg := <-messageChannel:
			logrus.Debug("Broadcast transmitted to network")
			checksum, err := checksum.GenerateJSONChecksum(msg.Payload)
			if err != nil {
				logrus.Error("Checksum generation failed:", err)
				continue
			}
			msg.Checksum = checksum
			msg.SenderId = nodeIP

			broadcastTransmissionChannel <- msg
		}
	}
}

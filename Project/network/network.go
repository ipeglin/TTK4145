package network

import (
	"elevator/checkpoint"
	"fmt"
	"network/broadcast"
	"network/checksum"
	"network/local"
	"network/nodes"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const lifelinePort int = 1337
const messagePort int = lifelinePort + 1

type Message struct {
	SenderId string // IPv4
	Payload  checkpoint.CombinedInput
	Checksum string
}

func Init(nodesChannel chan<- nodes.NetworkNodeRegistry, messageChannel <-chan Message, responseChannel chan<- Message, onlineStatusChannel chan<- bool, ipChannel chan<- string) {
	nodeIP, err := local.GetIP()
	if err != nil {
		logrus.Debug("Unable to get the IP address")
	}

	ipChannel <- nodeIP // pass the IP address to main process

	nodeUid := fmt.Sprintf("peer-%s-%d", nodeIP, os.Getpid())
	logrus.Debug(fmt.Sprintf("Network module initialised with UID=%s on PORT=%d", nodeUid, lifelinePort))

	// setup lifeline for network node registry
	nodeRegistryChannel := make(chan nodes.NetworkNodeRegistry)
	TransmissionEnableChannel := make(chan bool)
	go nodes.Sender(lifelinePort, nodeUid, TransmissionEnableChannel)
	go nodes.Receiver(lifelinePort, nodeRegistryChannel)

	// setup broadcast for message transmission
	broadcastTransmissionChannel := make(chan Message)
	broadcastReceiverChannel := make(chan Message)
	go broadcast.Sender(messagePort, broadcastTransmissionChannel)
	go broadcast.Receiver(nodeIP, messagePort, broadcastReceiverChannel)

	for {
		select {
		case reg := <-nodeRegistryChannel:
			logrus.Debug(fmt.Sprintf("Node registry update:\n  Nodes:    %q\n  New:      %q\n  Lost:     %q", reg.Nodes, reg.New, reg.Lost))

			// on state change, pass to main process
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

			sum, err := checksum.GenerateJSONChecksum(msg.Payload)
			if err != nil {
				logrus.Error("Checksum generation failed:", err)
				continue
			}
			logrus.Warn("Recieved checksum: ", msg.Checksum, "\nComputed checksum: ", sum)

			// drop incorrect payload
			if msg.Checksum != sum {
				logrus.Error("Checksum mismatch, payload corrupted. Abort forwarding")
				continue
			}

			responseChannel <- msg

		case msg := <-messageChannel:
			logrus.Debug("Broadcast transmitted to network")

			sum, err := checksum.GenerateJSONChecksum(msg.Payload)
			if err != nil {
				logrus.Error("Checksum generation failed:", err)
				continue
			}

			msg.SenderId = nodeIP
			msg.Checksum = sum
			logrus.Debug("Generated checksum: ", sum)

			broadcastTransmissionChannel <- msg
		}
	}
}

package messagehandler

import (
	"elevator/statehandler"
	"fmt"
	"messagehandler/checksum"
	"network/broadcast"
	"network/local"
	"network/nodes"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const lifelinePort int = 1337
const messagePort int = lifelinePort + 1
const createMessageInterval = 500

type Message struct {
	SenderId string // IPv4
	Payload  statehandler.ElevatorState
	Checksum string
}

func Init(nodesChannel chan<- nodes.NetworkNodeRegistry, responseChannel chan<- Message, onlineStatusChannel chan<- bool, ipChannel chan<- string) {
	nodeIP, err := local.GetIP()
	if err != nil {
		logrus.Debug("Unable to get the IP address")
	}

	ipChannel <- nodeIP

	nodeUid := fmt.Sprintf("peer-%s-%d", nodeIP, os.Getpid())
	logrus.Debug(fmt.Sprintf("Network module initialised with UID=%s on PORT=%d", nodeUid, lifelinePort))

	nodeRegistryChannel := make(chan nodes.NetworkNodeRegistry)
	TransmissionEnableChannel := make(chan bool)
	go nodes.Sender(lifelinePort, nodeUid, TransmissionEnableChannel)
	go nodes.Receiver(lifelinePort, nodeRegistryChannel)

	broadcastTransmissionChannel := make(chan Message)
	broadcastReceiverChannel := make(chan Message)
	go broadcast.Sender(messagePort, broadcastTransmissionChannel)
	go broadcast.Receiver(nodeIP, messagePort, broadcastReceiverChannel)

	messageChannel := make(chan Message)
	go createStateMessage(messageChannel)

	for {
		select {
		case reg := <-nodeRegistryChannel:
			if slices.Contains(reg.Lost, nodeUid) {
				logrus.Warn("Node lost connection:", nodeUid)
				onlineStatusChannel <- false
			} else if reg.New == nodeUid {
				logrus.Warn("Node connected:", nodeUid)
				onlineStatusChannel <- true
			}

			nodesChannel <- reg

		case msg := <-broadcastReceiverChannel:
			sum, err := checksum.GenerateJSONChecksum(msg.Payload)
			if err != nil {
				logrus.Error("Checksum generation failed:", err)
				continue
			}
			logrus.Debug("Checksum match: ", msg.Checksum == sum)

			if msg.Checksum != sum {
				logrus.Error("Checksum mismatch, payload corrupted. Abort forwarding")
				continue
			}

			logrus.Debug("Received message from ", msg.SenderId)
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

func createStateMessage(messageChannel chan Message) {
	for {
		elevState, _ := statehandler.LoadState()
		messageChannel <- Message{Payload: elevState}
		time.Sleep(createMessageInterval * time.Millisecond)
	}
}

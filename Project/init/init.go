package main

import (
	"elevator"
	"elevator/checkpoint"
	"elevator/fsm"
	"elevator/processpair"
	"logger"
	"network"
	"network/nodes"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func initNode(isPrimaryProcess bool) {
	var lostNodes []string
	var localStateFile string

	logger.Setup()
	logrus.Info("Node initialised with PID:", os.Getpid())

	nodeOverviewChannel := make(chan nodes.NetworkNodeRegistry)
	messageReceiveChannel := make(chan network.Message)
	messageTransmitterChannel := make(chan network.Message)
	onlineStatusChannel := make(chan bool)
	ipChannel := make(chan string)

	go network.Init(nodeOverviewChannel, messageTransmitterChannel, messageReceiveChannel, onlineStatusChannel, ipChannel)

	// await ip from network module
	localIP := <-ipChannel
	localStateFile = localIP + ".json"

	go elevator.Init(localIP, isPrimaryProcess)

	// broadcast state
	go func() {
		for {
			// TODO: If invalid json, do not broadcast, so ther nodes will think it is offline
			elv, _ := checkpoint.LoadCombinedInput(localStateFile)
			messageTransmitterChannel <- network.Message{Payload: elv}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// handle incoming messages
	for {
		select {
		case reg := <-nodeOverviewChannel:
			//hvis du går fra å være offline til online legges du ikke til. 
			// må fikses 
			
			logrus.Info("Known nodes:", reg.Nodes)
			if len(reg.Lost) <= 0 {
				lostNodes = []string{}
				continue
			}
			
			logrus.Warn("Lost nodes:", reg.Lost)

			// extract ip from node names
			var lostNodeAddresses []string
			for _, node := range reg.Lost {
				ip := strings.Split(node, "-")[1]
				lostNodeAddresses = append(lostNodeAddresses, ip)
			}
			logrus.Debug("Removing lost IPs: ", lostNodeAddresses)

			lostNodes = lostNodeAddresses // Update the lostNodes

			checkpoint.DeleteInactiveElevatorsFromJSON(lostNodes, localStateFile)
			fsm.JSONOrderAssigner(localStateFile, localIP)
			fsm.RequestButtonPressV3(localStateFile, localIP)

		case msg := <-messageReceiveChannel:
			// TODO: handle incoming messages
			logrus.Debug("Received message from ", msg.SenderId, ": ", msg.Payload)

			incomingState := msg.Payload
			// TODO: Reassign orders

			// update and remove list nodes
			if !checkpoint.IncomingDataIsCorrupt(incomingState) {
				//checkpoint.SaveCombinedInput(incomingState, incomingFileName)
				checkpoint.IncomingJSONHandling(localStateFile, incomingState, msg.SenderId)
				fsm.JSONOrderAssigner(localStateFile, localIP)
				fsm.RequestButtonPressV3(localStateFile, localIP) // TODO: Only have one version
			}

		case online := <-onlineStatusChannel:
			fsm.RebootJSON(localIP, localStateFile)
			fsm.JSONOrderAssigner(localStateFile, localIP)
			fsm.RequestButtonPressV3(localStateFile, localIP) // TODO: Only have one version
			logrus.Warn("Updated online status:", online)
		}
	}

}

func main() {
	var entryPointFunction processpair.TFunc = initNode
	processpair.ProcessPairHandler(entryPointFunction)

	// Block the main goroutine indefinitely
	done := make(chan struct{})
	<-done
}

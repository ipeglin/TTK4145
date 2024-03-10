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
	"time"

	"github.com/sirupsen/logrus"
)

func mainLogic(firstProcess bool) {
  logger.Setup()
	logrus.Info("Node initialised with PID:", os.Getpid())

	nodeOverviewChannel := make(chan nodes.NetworkNodeRegistry)
	messageReceiveChannel := make(chan network.Message)
	messageTransmitterChannel := make(chan network.Message)
	onlineStatusChannel := make(chan bool)
	ipChannel := make(chan string)

	go network.Init(nodeOverviewChannel, messageTransmitterChannel, messageReceiveChannel, onlineStatusChannel, ipChannel)

	localIP := <-ipChannel
	go elevator.Init(localIP, firstProcess)

	go func() {
		for {
			//antar det er her vi sender
			//dersom local elevator dedekteres ikke funksjonell ønsker vi ikke broacaste JSON
			//da vil alle andre heiser tro den er offline og ikke assigne den nye calls.
			localFilename := localIP + ".json"
			elv, _ := checkpoint.LoadCombinedInput(localFilename)
			messageTransmitterChannel <- network.Message{Payload: elv}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for {
		select {
		case reg := <-nodeOverviewChannel:
			logrus.Info("Known nodes:", reg.Nodes)
		case msg := <-messageReceiveChannel:
			//todo
			//load 			msg.Payload
			//
			//som får filnavn lik ip
			logrus.Info("Received message from ", msg.SenderId, ": ", msg.Payload)
			strings := make([]string, 8)
			// localIP
			// inncomigIP.JSON

			localFilname := localIP + ".json"
			incommigFilname := msg.SenderId + ".json"
			inncommingCombinedInput := msg.Payload

			//her må vi reassigne
			//temp solution
			//NOT VERY NICE. ONLY PROOF OF CONCEPT
			if !checkpoint.IncomingDataIsCorrupt(inncommingCombinedInput) {
				checkpoint.InncommingJSONHandeling(localFilname, incommigFilname, inncommingCombinedInput, strings)
				fsm.FsmJSONOrderAssigner(localFilname, localIP)
				fsm.FsmRequestButtonPressV3(localFilname, localIP)
			}

		case online := <-onlineStatusChannel:
			logrus.Warn("Updated online status:", online)
		}
	}

}

func main() {
	var mainFuncObject processpair.MainFuncType = mainLogic
	processpair.ProcessPairHandler(mainFuncObject)

	// Block the main goroutine indefinitely
	done := make(chan struct{})
	<-done
}

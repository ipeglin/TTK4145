package main

import (
	"elevator"
	"elevator/checkpoint"
	"elevator/fsm"
	"elevator/processpair"
	"fmt"
	"logger"
	"network"
	"network/nodes"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func mainLogic(firstProcess bool) {
	var lostNodes []string

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
			localFilname := localIP + ".json"
			lostNodes = reg.Lost
			logrus.Info("Known nodes:", reg.Nodes)
			var updatedLostNodes []string // This will hold the processed IP addresses
			if len(reg.Lost) > 0 {
				logrus.Info("Lost nodes:", reg.Lost)
				// Handling lost nodes
				for _, lostIP := range reg.Lost {
					parts := strings.Split(lostIP, "-")
					if len(parts) >= 3 { // Ensure the format matches "Peer-IP-SOME_NUMBERS"
						ip := parts[1] // The IP address is the second element
						updatedLostNodes = append(updatedLostNodes, ip)
						logrus.Info("Processing lost node IP:", ip)
						//print(ip)
					}
				}
				lostNodes = updatedLostNodes // Update the lostNodes with just the IPs
				for _, id := range lostNodes {
					fmt.Println(id) // Using fmt.Println for printing each ID on a new line
				}
				checkpoint.DeleteInactiveElevatorsFromJSON(lostNodes, localFilname)
				fsm.FsmJSONOrderAssigner(localFilname, localIP)
				fsm.FsmRequestButtonPressV3(localFilname, localIP)
			}

		case msg := <-messageReceiveChannel:
			//todo
			//load 			msg.Payload
			//
			//som får filnavn lik ip
			logrus.Info("Received message from ", msg.SenderId, ": ", msg.Payload)
			//strings := make([]string, 8)
			// localIP
			// inncomigIP.JSON

			localFilname := localIP + ".json"
			incommigFilname := msg.SenderId + ".json"
			inncommingCombinedInput := msg.Payload
			//print(lostNodes)
			//her må vi reassigne
			//temp solution
			//NOT VERY NICE. ONLY PROOF OF CONCEPT
			//print(lostNodes)
			if !checkpoint.IncomingDataIsCorrupt(inncommingCombinedInput) {
				//oppdater og fjern lostNodes
				checkpoint.InncommingJSONHandeling(localFilname, incommigFilname, inncommingCombinedInput, lostNodes)
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

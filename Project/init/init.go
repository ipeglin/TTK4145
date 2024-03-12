package main

import (
	"elevator"
	"elevator/fsm"
	"elevator/jsonhandler"
	"logger"
	"network"
	"network/nodes"
	"os"
	"processpair"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func initNode(isFirstProcess bool) {
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

	go elevator.Init(localIP, isFirstProcess)

	// broadcast state
	go func() {
		for {
			// TODO: If invalid json, do not broadcast, so ther nodes will think it is offline
			elv, _ := jsonhandler.LoadCombinedInput(localStateFile)
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
			if len(reg.Lost) > 0 {
				logrus.Warn("Lost nodes:", reg.Lost)
			}

			// extract ip from node names
			var lostNodeAddresses []string
			for _, node := range reg.Lost {
				ip := strings.Split(node, "-")[1]
				lostNodeAddresses = append(lostNodeAddresses, ip)
			}
			logrus.Debug("Removing lost IPs: ", lostNodeAddresses)

			jsonhandler.DeleteInactiveElevatorsFromJSON(lostNodeAddresses, localStateFile)
			if fsm.OnlyElevatorOnlie(localStateFile, localIP) {
				fsm.JSONOrderAssigner(localStateFile, localIP)
				jsonhandler.JSONsetAllLights(localStateFile, localIP)
				fsm.MoveOnActiveOrders(localStateFile, localIP)
			}

			//skal vi reasigne her? nei?
			//dersom vi ikke og den er enset igjen online så vil den ta alle den har blitt assignet (kan være mer enn en og fuløre dem)
			//fsm.JSONOrderAssigner(localStateFile, localIP)
			//fsm.MoveOnActiveOrders(localStateFile, localIP)

		case msg := <-messageReceiveChannel:
			// TODO: handle incoming messages
			logrus.Debug("Received message from ", msg.SenderId)

			incomingState := msg.Payload
			// TODO: Reassign orders

			// update and remove list nodes
			if !jsonhandler.IncomingDataIsCorrupt(incomingState) {
				fsm.HandleIncomingJSON(localStateFile, localIP, msg.Payload, msg.SenderId)
				fsm.AssignOnInncoming(localStateFile, localIP,msg.Payload)
				fsm.MoveOnActiveOrders(localStateFile, localIP)
				
				//fsm.UpdateElevatorState(localIP, localStateFile)

				//fsm.HandleIncomingJSON(localStateFile, incomingState, msg.SenderId)
				//checkpoint.JSONsetAllLights(localStateFile, msg.SenderId)
				fsm.JSONOrderAssigner(localStateFile, localIP)
				// ! Only have one version
			}

		case online := <-onlineStatusChannel:
			//ViErOnline = True
			fsm.HandleStateOnReboot(localIP, localStateFile) // Deprecated: fsm.RebootJSON()
			//fsm.JSONOrderAssigner(localStateFile, localIP)
			//fsm.MoveOnActiveOrders(localStateFile, localIP) // ! Only have one version
			logrus.Warn("Updated online status:", online)
		}
		//case ofline :=<-
		//ViErOnline = False
	}

}

func main() {
	var entryPointFunction processpair.TFunc = initNode
	processpair.CreatePair(entryPointFunction)

	// Block the main goroutine indefinitely
	done := make(chan struct{})
	<-done
}

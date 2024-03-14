package main

import (
	"elevator"
	"elevator/fsm"
	"elevator/statehandler"
	"logger"
	"network"
	"network/nodes"
	"os"
	"processpair"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Todo ! : Ikke kall init node kanskje?
func initNode(isFirstProcess bool) {
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
	go elevator.Init(localIP, isFirstProcess)

	// broadcast state
	go func() {
		for {
			// TODO: If invalid json, do not broadcast, so ther nodes will think it is offline
			elv, _ := statehandler.LoadState()
			messageTransmitterChannel <- network.Message{Payload: elv}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for {
		select {
		case reg := <-nodeOverviewChannel:
			logrus.Info("Known nodes:", reg.Nodes)
			if len(reg.Lost) > 0 {
				logrus.Warn("Lost nodes:", reg.Lost)
			}
			// extract ip from node names
			//TODO: Phillip kan du lage sånn at nodene bare inneholder ip
			var lostNodeAddresses []string
			for _, node := range reg.Lost {
				ip := strings.Split(node, "-")[1]
				lostNodeAddresses = append(lostNodeAddresses, ip)
			}
			logrus.Debug("Removing lost IPs: ", lostNodeAddresses)

			statehandler.RemoveElevatorsFromJSON(lostNodeAddresses)
			if statehandler.IsOnlyNodeOnline(localIP) {
				fsm.AssignOrders(localIP)
				fsm.SetConfirmedHallLights(localIP)
				fsm.MoveOnActiveOrders(localIP)
				fsm.UpdateElevatorState(localIP)

			}

		case msg := <-messageReceiveChannel:
			logrus.Debug("Received message from ", msg.SenderId)
			//if !statehandler.IsStateCorrupted(msg.Payload) {
			statehandler.HandleIncomingSate(localIP, msg.Payload, msg.SenderId)
			fsm.AssignIfWorldViewsAlign(localIP, msg.Payload)
			fsm.MoveOnActiveOrders(localIP)
			fsm.UpdateElevatorState(localIP)

		case online := <-onlineStatusChannel:
			if online {
				fsm.HandleStateOnReboot(localIP)
			} else {
				fsm.SetAllLights()
				fsm.MoveOnActiveOrders(localIP)
			}
			logrus.Warn("Updated online status:", online)
		}
	}

}

//Todo: ! Kan vi omdøpe over til å bli main og calle diise func i main

func main() {
	var entryPointFunction func(bool) = initNode
	processpair.CreatePair(entryPointFunction)

	// Block the main goroutine indefinitely
	done := make(chan struct{})
	<-done
}

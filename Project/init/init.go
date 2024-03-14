package main

import (
	"elevator"
	"elevator/fsm"
	"elevator/statehandler"
	"logger"
	"messagehandler"
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
	messageReceiveChannel := make(chan messagehandler.Message)
	messageTransmitterChannel := make(chan messagehandler.Message)
	onlineStatusChannel := make(chan bool)
	ipChannel := make(chan string)

	go messagehandler.Init(nodeOverviewChannel, messageTransmitterChannel, messageReceiveChannel, onlineStatusChannel, ipChannel)

	// await ip from network module
	localIP := <-ipChannel
	go elevator.Init(localIP, isFirstProcess)

	// broadcast state
	go func() {
		for {
			elv, _ := statehandler.LoadState()
			messageTransmitterChannel <- messagehandler.Message{Payload: elv}
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
			var lostNodeAddresses []string
			for _, node := range reg.Lost {
				ip := strings.Split(node, "-")[1]
				lostNodeAddresses = append(lostNodeAddresses, ip)
			}
			logrus.Debug("Removing lost IPs: ", lostNodeAddresses)

			statehandler.RemoveElevatorsFromState(lostNodeAddresses)
			if statehandler.IsOnlyNodeOnline(localIP) {
				fsm.AssignOrders(localIP)
				fsm.SetConfirmedHallLights(localIP)
				fsm.MoveOnActiveOrders(localIP)
				fsm.UpdateElevatorState(localIP)
			}

		case msg := <-messageReceiveChannel:
			logrus.Debug("Received message from ", msg.SenderId)
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

func main() {
	var entryPointFunction func(bool) = initNode
	processpair.CreatePair(entryPointFunction)

	done := make(chan struct{})
	<-done
}

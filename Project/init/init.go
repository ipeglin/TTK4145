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

	"github.com/sirupsen/logrus"
)

func initNode(isFirstProcess bool) {
	logger.Setup()
	logrus.Info("Node initialised with PID:", os.Getpid())

	nodeOverviewChannel := make(chan nodes.NetworkNodeRegistry)
	messageReceiveChannel := make(chan messagehandler.Message)
	onlineStatusChannel := make(chan bool)
	ipChannel := make(chan string)
	go messagehandler.Init(nodeOverviewChannel, messageReceiveChannel, onlineStatusChannel, ipChannel)

	localIP := <-ipChannel
	go elevator.Init(localIP, isFirstProcess)

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
		}
	}
}

func main() {
	var entryPointFunction func(bool) = initNode
	processpair.CreatePair(entryPointFunction)

	done := make(chan struct{})
	<-done
}

package main

import (
	"elevator"
	"elevator/elevatorcontroller"
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

const initWaitTimeMS = 100

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
	time.Sleep(initWaitTimeMS * time.Millisecond)

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
				elevatorcontroller.AssignOrders(localIP)
				elevatorcontroller.SetConfirmedHallLights(localIP)
				elevatorcontroller.MoveOnActiveOrders(localIP)
				elevatorcontroller.UpdateElevatorState(localIP)
			}

		case msg := <-messageReceiveChannel:
			if !statehandler.IsStateCorrupted(msg.Payload) {
				statehandler.HandleIncomingSate(localIP, msg.Payload, msg.SenderId)
				elevatorcontroller.AssignIfWorldViewsAlign(localIP, msg.Payload)
				elevatorcontroller.MoveOnActiveOrders(localIP)
			}
			elevatorcontroller.UpdateElevatorState(localIP)
		case online := <-onlineStatusChannel:
			if online {
				elevatorcontroller.HandleStateOnReboot(localIP)
			} else {
				elevatorcontroller.SetAllLights()
				elevatorcontroller.MoveOnActiveOrders(localIP)
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

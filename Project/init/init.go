package main

import (
	"elevator"
	"elevator/checkpoint"

	//"fmt"

	"network"
	"network/nodes"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Node initialised with PID:", os.Getpid())

	nodeOverviewChannel := make(chan nodes.NetworkNodeRegistry)
	messageReceiveChannel := make(chan network.Message)
	messageTransmitterChannel := make(chan network.Message)
	onlineStatusChannel := make(chan bool)
	ipChannel := make(chan string)

	go network.Init(nodeOverviewChannel, messageTransmitterChannel, messageReceiveChannel, onlineStatusChannel, ipChannel)

	// TODO: Launch new process watching current process in case of crash
	localIP := <-ipChannel
	go elevator.Init(localIP)

	go func() {
		for {
			elv, _ := checkpoint.LoadCombinedInput(checkpoint.JSONFile)
			messageTransmitterChannel <- network.Message{Payload: elv, MessageId: 0}
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
			strings := make([] string, 0)
			// localIP
			// inncomigIP.JSON 
			localFilname := localIP + ".JSON"
			incommigFilname := msg.SenderId +".JSON"
			inncommingCombinedInput := msg.Payload
			checkpoint.InncommingJSONHandeling(localFilname , incommigFilname , inncommingCombinedInput, strings)
		case online := <-onlineStatusChannel:
			logrus.Warn("Updated online status:", online)
		}
	}
}

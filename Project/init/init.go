package main

import (
	"elevator"
	"elevator/checkpoint"
	"fmt"
	"network"
	"network/nodes"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
  // get project root path
  projectRoot, err := filepath.Abs("../")
  if err != nil {
      logrus.Fatal("Failed to find project root")
  }

  // generate log name
  now := time.Now()
  logFile := fmt.Sprintf("runtime_%d-%d-%d_%d:%d:%d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Hour(),
		now.Second()
  )

  // pass log file to logrus
  f, err := os.OpenFile(projectRoot + "/log/" + logFile + ".log", os.O_WRONLY | os.O_CREATE, 0755)
  if err != nil {
      # handle
  }
  logrus.SetOutput(f)
}

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
			//antar det er her vi sender
			//dersom local elevator dedekteres ikke funksjonell ønsker vi ikke broacaste JSON
			//da vil alle andre heiser tro den er offline og ikke assigne den nye calls.
			localFilname := localIP + ".json"
			elv, _ := checkpoint.LoadCombinedInput(localFilname)
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
			incomingFilename := msg.SenderId + ".json"
			incomingCombinedInput := msg.Payload
			checkpoint.IncomingJSONHandeling(localFilname, incomingFilename, incomingCombinedInput, strings)
		case online := <-onlineStatusChannel:
			logrus.Warn("Updated online status:", online)
		}
	}
}

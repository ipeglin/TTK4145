package processpair

import (
	"elevator/checkpoint"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

var timeLimitOnline time.Duration = time.Duration(3 * time.Second)

type TFunc func(bool)

func CreatePair(f TFunc) {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		logrus.Info("Initiated as backup process. Listening for main process to terminate...")
		waitForProcessTermination(f)
	} else {
		startProcess(f, true)
	}
}

func waitForProcessTermination(f TFunc) {
	for {
		if !isProcessAlive() {
			startProcess(f, false)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// ? Maybe pass this is as param to CreatePair to detach module from elevator/checkpoint
func isProcessAlive() bool {
	_, checkpointTime, _ := checkpoint.LoadCheckpoint()
	timeSinceCheckpoint := time.Since(checkpointTime)
	return timeLimitOnline >= timeSinceCheckpoint
}

func startProcess(mainFunc TFunc, firstProcsess bool) {
	logrus.Info("Main process initiated...")
	go mainFunc(firstProcsess)
	startBackup()
}

func startBackup() {
	cmd := exec.Command("gnome-terminal", "--", "./elevator", "backup")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start backup process in a new terminal: %s\n", err)
	} else {
		fmt.Println("Backup process started successfully in a new terminal")
	}
}

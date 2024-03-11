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

func startMainProcess(mainFunc TFunc, firstProcsess bool) {
	logrus.Info("Main process initiated")
	go mainFunc(firstProcsess)
	startBackupProcess()

}

func startBackupProcess() {
	cmd := exec.Command("gnome-terminal", "--", "./elevator", "backup")
	if err := cmd.Start(); err != nil {
		logrus.Error(fmt.Sprintf("Failed to start backup process in a new terminal: %s\n", err))
	} else {
		logrus.Info("Backup process started successfully in a new terminal")
	}
}

func monitorMainProcessAndTakeOver(mainFunc TFunc) {
	for {
		if !isMainAlive() {
			startMainProcess(mainFunc, false)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func isMainAlive() bool {
	_, checkpointTime, _ := checkpoint.LoadElevatorCheckpoint(checkpoint.FilenameCheckpoint)
	timeSinceCheckpoint := time.Since(checkpointTime)
	return timeLimitOnline >= timeSinceCheckpoint
}

func ProcessPairHandler(mainFunc TFunc) {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		logrus.Info("Backup initiated. Listing for main process to terminate...")
		monitorMainProcessAndTakeOver(mainFunc)
	} else {
		startMainProcess(mainFunc, true)
	}
}

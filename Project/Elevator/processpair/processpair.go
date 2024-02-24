package processpair

import (
	"fmt"
	"heislab/Elevator/checkpoint"
	"os"
	"os/exec"
	"time"
)

var timeLimitOnline time.Duration = time.Duration(2000 * time.Millisecond)

type MainFuncType func()

func startMainProcess(mainFunc MainFuncType) {
	print("Im main bitch")
	go mainFunc()
	startBackupProcess()

}

/*func startBackupProcess() {
	cmd := exec.Command(os.Args[0], "backup")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start backup process: %s\n", err)
	} else {
		fmt.Println("Backup process started successfully")
	}
}*/

func startBackupProcess() {
	cmd := exec.Command("gnome-terminal", "--", "./myElevatorProgram", "backup")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start backup process in a new terminal: %s\n", err)
	} else {
		fmt.Println("Backup process started successfully in a new terminal")
	}
}

func monitorMainProcessAndTakeOver(mainFunc MainFuncType) {
	for {
		if !isMainAlive() {
			startMainProcess(mainFunc)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func isMainAlive() bool {
	_, checkpointTime, _ := checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
	timeSinceCheckpoint := time.Since(checkpointTime)
	return timeLimitOnline >= timeSinceCheckpoint
}

func ProcessPairHandler(mainFunc MainFuncType) {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		print("Im backup")
		monitorMainProcessAndTakeOver(mainFunc)
	} else {
		startMainProcess(mainFunc)
	}
}

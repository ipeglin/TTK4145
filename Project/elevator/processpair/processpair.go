package processpair

import (
	"elevator/checkpoint"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var timeLimitOnline time.Duration = time.Duration(2000 * time.Millisecond)

type MainFuncType func(bool)

func startMainProcess(mainFunc MainFuncType, firstProcsess bool) {
	print("Im main bitch")
	go mainFunc(firstProcsess)
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
			startMainProcess(mainFunc, false)
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
		startMainProcess(mainFunc, true)
	}
}

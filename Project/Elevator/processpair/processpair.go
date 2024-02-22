package processpair

import (
	"fmt"
	"os"
	"os/exec"
)

type MainFuncType func()

func startMainProcess(mainFunc MainFuncType) {
	go mainFunc()
	startBackupProcess()

}

func startBackupProcess() {
	cmd := exec.Command(os.Args[0], "backup")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start backup process: %s\n", err)
	} else {
		fmt.Println("Backup process started successfully")
	}
}

func monitorMainProcessAndTakeOver(mainFunc MainFuncType) {
	if isMainAlive() != true {
		startMainProcess(mainFunc)
	}
}

func isMainAlive() bool {

	return true
}

func ProcessPairHandler(mainFunc MainFuncType) {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		monitorMainProcessAndTakeOver(mainFunc)
	} else {
		startMainProcess(mainFunc)
	}
}

func ProcessPairCheckpoint(data []byte) {
	//Må skrive tid + alle elevator states i en fil
}

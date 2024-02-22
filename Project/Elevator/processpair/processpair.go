package processpair

import (
	"os"
)

type MainFuncType func()

func StartMainProcess(mainFunc MainFuncType) {
	go mainFunc()

}

func MonitorMainProcess(mainFunc MainFuncType) {

}

func TakeOverMainProcess(mainFunc MainFuncType) {
	if isMainAlive() != true {
		StartMainProcess(mainFunc)
	}
}

func isMainAlive() bool {

	return true
}

func ProcessPairHandler(mainFunc MainFuncType) {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		MonitorMainProcess(mainFunc)
	} else {
		StartMainProcess(mainFunc)
	}
}

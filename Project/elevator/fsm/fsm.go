package fsm

import (
	"elevator/checkpoint"
	"elevator/elev"
	"elevator/elevio"
	"elevator/requests"
	"elevator/timer"
	"fmt"
	"network/local"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var elevator elev.Elevator
var outputDevice elevio.ElevOutputDevice
var elevatorName string
var localStateFile string

func init() {
	elevator = elev.ElevatorInit()
	elevatorName, _ = local.GetIP()
	outputDevice = elevio.ElevioGetOutputDevice()

	localStateFile = elevatorName + ".json"

	//Burde dette gå et annet sted?
	setAllLights()
	elevio.RequestDoorOpenLamp(false)
	elevio.RequestStopLamp(false)
}

// BUG: Init and SetElevator crashes when using process pairs
func SetElevator(f int, cb elev.ElevatorBehaviour, dirn elevio.ElevDir, r [elevio.NFloors][elevio.NButtons]bool, c elev.ElevatorConfig) {
	elevator.CurrentFloor = f
	elevator.CurrentBehaviour = cb
	elevator.Dirn = dirn
	elevator.Requests = r
	elevator.Config = c
}

func setAllLights() {
	for floor := 0; floor < elevio.NFloors; floor++ {
		outputDevice.RequestButtonLight(floor, elevio.BCab, elevator.Requests[floor][elevio.BCab])
	}
}

func MoveDownToFloor() {
	dirn := elevio.DirDown
	outputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

func FloorArrival(newFloor int, elevatorName string, filename string) {
	logrus.Warn("Arrived at new floor: ", newFloor)
	//elev.ElevatorPrint(elevator)
	elevator.CurrentFloor = newFloor
	outputDevice.FloorIndicator(elevator.CurrentFloor)
	//Helt unødvendig med switch her?
	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.RequestsShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			outputDevice.DoorLight(true)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			setAllLights()
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	}
	//fmt.Println("New state:")
	//elev.ElevatorPrint(elevator)
}

func FsmDoorTimeout(filename string, elevatorName string) {
	//fmt.Printf("\n\n%s()\n", "FsmDoorTimeout")
	//elev.ElevatorPrint(elevator)
	//Hvorfor switch
	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour

		switch elevator.CurrentBehaviour {
		case elev.EBDoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)
			setAllLights()

		case elev.EBMoving:
			outputDevice.DoorLight(false)
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevio.DirStop), " in FsmDoorTimeout")
			outputDevice.MotorDirection(elevator.Dirn)
		case elev.EBIdle:
			outputDevice.DoorLight(false)
		}

	}
	//fmt.Println("New State: ")
	//elev.ElevatorPrint(elevator)
}

func FsmObstruction() {
	if !timer.TimerInf {
		timer.TimerStartInf()
		if elevator.CurrentBehaviour == elev.EBIdle {
			outputDevice.DoorLight(true)
			elevator.CurrentBehaviour = elev.EBDoorOpen
		}
	} else {
		timer.TimerStopInf()
		timer.TimerStart(elevator.Config.DoorOpenDurationS)
	}
}

func FsmMakeCheckpointGo() {
	for {
		checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
		time.Sleep(50 * time.Millisecond)
	}

}

func FsmMakeCheckpoint() {
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func FsmResumeAtLatestCheckpoint(floor int) {
	elevator, _, _ = checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
	setAllLights()
	//fmt.Print(elevator.Dirn)
	if elevator.Dirn != elevio.DirStop && floor == -1 {
		outputDevice.MotorDirection(elevator.Dirn)
	}
	if floor != -1 {
		timer.TimerStart(elev.DoorOpenDurationSConfig)
		outputDevice.DoorLight(true)
	}
}

func FsmLoadLatestCheckpoint() {
	elevator, _, _ = checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
}

// Json fra her
func FsmInitJson(filename string, ElevatorName string) {
	// Gjør endringer på combinedInput her
	print(filename)
	err := os.Remove(filename)
	if err != nil {
		fmt.Println("Feil ved fjerning:", err)
	}
	combinedInput := checkpoint.InitializeCombinedInput(elevator, ElevatorName)

	// If the file was successfully deleted, return nil
	err = checkpoint.SaveCombinedInput(combinedInput, filename)
	if err != nil {
		fmt.Println("Feil ved lagring:", err)
	}
}

func FsmUpdateJSON(elevatorName string, filename string) {
	checkpoint.UpdateJSON(elevator, filename, elevatorName)
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func FsmRebootJSON(elevatorName string, filename string) {
	checkpoint.RebootJSON(elevator, filename, elevatorName)
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
}

func fsmUpdateJSONWhenNewOrderOccurs(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	checkpoint.UpdateJSONWhenNewOrderOccurs(filename, elevatorName, btnFloor, btn)
}

func FsmJSONOrderAssigner(filename string, elevatorName string) {
	checkpoint.JSONOrderAssigner(&elevator, filename, elevatorName)
}

func FsmRequestButtonPressV2(btnFloor int, btn elevio.Button, elevatorName string, filename string) {
	if requests.RequestsShouldClearImmediately(elevator, btnFloor, btn) && (elevator.CurrentBehaviour == elev.EBDoorOpen) {
		timer.TimerStart(elevator.Config.DoorOpenDurationS)
	} else {
		//elevator.Requests[btnFloor][btn] = true
		//trenger å sjekke at alt dette er riktig
		fsmUpdateJSONWhenNewOrderOccurs(btnFloor, btn, elevatorName, filename)
		print("funksjonskall funker")
		if btn == elevio.BCab{
			print("hei")
			elevator.Requests[btnFloor][btn] = true
		}
		//fsmJSONOrderAssigner(filename, elevatorName)
	}
}

// etter denne func broadcaster vi.
// så assigner vi
// så kaller vi denne
func FsmRequestButtonPressV3(filename string, elevatorName string) {
	switch elevator.CurrentBehaviour {
	case elev.EBIdle:
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator, filename, elevatorName)

		case elev.EBMoving:
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevator.Dirn), " in FsmRequestButtonPress")
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights()
	//fmt.Printf("New state: \n")
	//elev.ElevatorPrint(elevator)
}


func InncommingJSONHandeling(localFilname string, localElevatorName string, otherCombinedInput checkpoint.CombinedInput, incomingElevatorName string) {
	localCombinedInput, _ := checkpoint.LoadCombinedInput(localFilname)
	allValuesEqual := true
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] != localCombinedInput.CyclicCounter.HallRequests[f][i] {
				allValuesEqual = false
				break
			}
		}
	}

	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] > localCombinedInput.CyclicCounter.HallRequests[f][i] {
				localCombinedInput.CyclicCounter.HallRequests[f][i] = otherCombinedInput.CyclicCounter.HallRequests[f][i]
				localCombinedInput.HRAInput.HallRequests[f][i] = otherCombinedInput.HRAInput.HallRequests[f][i]
			}
			if (otherCombinedInput.CyclicCounter.HallRequests[f][i] == localCombinedInput.CyclicCounter.HallRequests[f][i]){ 
			 	if (localCombinedInput.HRAInput.HallRequests[f][i] != otherCombinedInput.HRAInput.HallRequests[f][i]){
					//midliertilig konflikt logikk dersom den ene er true og den andre er false 
					//oppstår ved motostop og bostruksjoner etc dersom den har selv claimet en orde som blir utført ila den har motorstop
					//Tenk om dette er beste løsning  
					localCombinedInput.HRAInput.HallRequests[f][i] = false
				} 
			}
		}
	}
	if _, exists := otherCombinedInput.HRAInput.States[incomingElevatorName]; exists {
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; !exists {
			localCombinedInput.HRAInput.States[incomingElevatorName] = otherCombinedInput.HRAInput.States[incomingElevatorName]
			localCombinedInput.CyclicCounter.States[incomingElevatorName] = otherCombinedInput.CyclicCounter.States[incomingElevatorName]
		} else {
			if otherCombinedInput.CyclicCounter.States[incomingElevatorName] > localCombinedInput.CyclicCounter.States[incomingElevatorName] {
				localCombinedInput.HRAInput.States[incomingElevatorName] = otherCombinedInput.HRAInput.States[incomingElevatorName]
				localCombinedInput.CyclicCounter.States[incomingElevatorName] = otherCombinedInput.CyclicCounter.States[incomingElevatorName]
			}
		}
	}else{
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; exists {
			delete(localCombinedInput.HRAInput.States, incomingElevatorName)
			delete(localCombinedInput.CyclicCounter.States, incomingElevatorName)
		}
	}
	if _, exists := otherCombinedInput.CyclicCounter.States[localElevatorName]; exists {
		if otherCombinedInput.CyclicCounter.States[localElevatorName] > localCombinedInput.CyclicCounter.States[localElevatorName] {
			localCombinedInput.CyclicCounter.States[localElevatorName] = otherCombinedInput.CyclicCounter.States[localElevatorName] +1 
		}
	}
	if allValuesEqual {
		// Execute further actions here
		checkpoint.JSONsetAllLights(localFilname, localElevatorName)
		checkpoint.JSONOrderAssigner(& elevator, localFilname, localElevatorName)
		//fsm.FsmRequestButtonPressV3(localFilname, localElevatorName) // TODO: Only have one version
	}
	FsmRequestButtonPressV3(localFilname, localElevatorName)
	checkpoint.SaveCombinedInput(localCombinedInput, localFilname)
}
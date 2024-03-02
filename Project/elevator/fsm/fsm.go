package fsm

import (
	"elevator/checkpoint"
	"elevator/driver/hwelevio"
	"elevator/elev"
	"elevator/elevio"
	"elevator/requests"
	"elevator/timer"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

var elevator elev.Elevator
var outputDevice elevio.ElevOutputDevice

func init() {
	elevator = elev.ElevatorInit()
	//fmt.Println("fsm_init has happend")
	//TODO
	outputDevice = elevio.ElevioGetOutputDevice()
	//Burde dette gå et annet sted?
	setAllLights()
	hwelevio.SetDoorOpenLamp(false)
	hwelevio.SetStopLamp(false)
}

// init og denne vil kræsje ved process pair må finnes ut av
func SetElevator(f int, cb elev.ElevatorBehaviour, dirn elevio.ElevDir, r [elevio.NFloors][elevio.NButtons]bool, c elev.ElevatorConfig) {
	elevator.CurrentFloor = f
	elevator.CurrentBehaviour = cb
	elevator.Dirn = dirn
	elevator.Requests = r
	elevator.Config = c
}

func setAllLights() {
	for floor := 0; floor < elevio.NFloors; floor++ {
		for btn := elevio.BHallUp; btn <= elevio.BCab; btn++ {
			outputDevice.RequestButtonLight(floor, btn, elevator.Requests[floor][btn])
			//fmt.Println(floor, " ", hwelevio.ButtonToString(btn), " ", elevator.Requests[floor][btn])
		}
	}
}

func FsmInitBetweenFloors() {
	dirn := elevio.DirDown
	outputDevice.MotorDirection(dirn)
	elevator.Dirn = dirn
	elevator.CurrentBehaviour = elev.EBMoving
}

func FsmRequestButtonPress(btnFloor int, btn elevio.Button) {

	//fmt.Printf("\n\n%s(%d, %s)\n", "FsmRequestButtonPress", btnFloor, elevio.ButtonToString(btn))
	//elev.ElevatorPrint(elevator)

	switch elevator.CurrentBehaviour {
	case elev.EBDoorOpen:
		if requests.RequestsShouldClearImmediately(elevator, btnFloor, btn) {
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
		} else {
			elevator.Requests[btnFloor][btn] = true
		}

	case elev.EBMoving:
		elevator.Requests[btnFloor][btn] = true

	case elev.EBIdle:
		elevator.Requests[btnFloor][btn] = true
		pair := requests.RequestsChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.CurrentBehaviour = pair.Behaviour
		switch pair.Behaviour {
		case elev.EBDoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)

		case elev.EBMoving:
			//fmt.Println("Calling MotorDirection: ", elevio.ElevDirToString(elevator.Dirn), " in FsmRequestButtonPress")
			outputDevice.MotorDirection(elevator.Dirn)
		}
	}
	setAllLights()
	//fmt.Printf("New state: \n")
	//elev.ElevatorPrint(elevator)
}

func FsmFloorArrival(newFloor int) {
	//fmt.Printf("\n\n%s(%d)\n", "FsmFloorArrival", newFloor)
	//elev.ElevatorPrint(elevator)
	elevator.CurrentFloor = newFloor
	outputDevice.FloorIndicator(elevator.CurrentFloor)
	//Helt unødvendig med switch her?
	switch elevator.CurrentBehaviour {
	case elev.EBMoving:
		if requests.RequestsShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.DirStop)
			//elevator.Dirn = elevio.DirStop
			outputDevice.DoorLight(true)
			elevator = requests.RequestsClearAtCurrentFloor(elevator)
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			setAllLights()
			elevator.CurrentBehaviour = elev.EBDoorOpen
			checkpoint.CallCompleteToJSON("one",checkpoint.FilenameHRAInput, elevator)
		}
	}
	//fmt.Println("New state:")
	//elev.ElevatorPrint(elevator)
}

func FsmDoorTimeout() {
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
			elevator = requests.RequestsClearAtCurrentFloor(elevator)
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

// TODO
func FsmObstruction() {
	if elevator.CurrentBehaviour == elev.EBDoorOpen {
		timer.TimerStop()
		timer.TimerStart(elevator.Config.DoorOpenDurationS)
		//fmt.Print("timer started")
	}

}

/*
// TODO
// Huske state før stop, så resume den? Tror det vil være en god løsning, midlertidig løsning for nå
func FsmStop(stop bool) {
	FsmMakeCheckpoint()
	fmt.Print("kallet stopp: ", stop)
	outputDevice.StopButtonLight(stop)
	if stop {
		elevator.Dirn = elevio.DirStop
		outputDevice.MotorDirection(elevator.Dirn)
		if elevio.InputDevice.FloorSensor() != -1 {
			elevator.CurrentBehaviour = elev.EBDoorOpen
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			hwelevio.SetDoorOpenLamp(true)
		}
	} else {
		FsmResumeAtLatestCheckpoint()
	}
}

/*
func FsmStop(stop bool) {
	fmt.Println("FsmStop(): ", stop)
	elev.ElevatorPrint(elevator)
	hwelevio.SetStopLamp(stop)
	if stop {
		requests.RequestsClearAll(&elevator)
		setAllLights()
		outputDevice.MotorDirection(elevio.ElevDir(hwelevio.MD_Stop))
		elevator.Dirn = elevio.DirStop
		if elevio.InputDevice.FloorSensor() != -1 {
			elevator.CurrentBehaviour = elev.EBDoorOpen
			timer.TimerStart(elevator.Config.DoorOpenDurationS)
			hwelevio.SetDoorOpenLamp(true)
		}
	} else {
		if elevator.CurrentBehaviour == elev.EBDoorOpen {
			elevator.CurrentBehaviour = elev.EBIdle
		} else if elevio.InputDevice.FloorSensor() != -1 {
			FsmInitBetweenFloors()
		}
	}
	elev.ElevatorPrint(elevator)
}*/
/*
func FsmMakeCheckpoint() {
	checkpoint.SaveElevCheckpoint(elevator, checkpoint.FilenameCheckpoint)
	//fmt.Print("The elevator which were saved: \n")
	//elev.ElevatorPrint(elevator)
}

func FsmResumeAtLatestCheckpoint() {
	elevator, _, _ = checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
	//fmt.Print(elevator.Dirn)
	outputDevice.MotorDirection(elevator.Dirn)
}

func FsmLoadLatestCheckpoint() {
	elevator, _, _ = checkpoint.LoadElevCheckpoint(checkpoint.FilenameCheckpoint)
}

func FsmTestProcessPair() {
	for {
		FsmLoadLatestCheckpoint()
		elevator.CurrentFloor += 1
		fmt.Print(elevator.CurrentFloor)
		FsmMakeCheckpoint()
		time.Sleep(1000 * time.Millisecond)
	}

}
*/

// ///AV GUSTAV TESTER OG FUCKER
//////NB HUSK Å BYTTE one med id til heis 


//her er det for å oppdatere etter buttoncall 
func FsmUpdataJSONOnbtnEvent() {
	elevatorName := "one"     // Example elevator name, adjust as necessary.
	localElevator := elevator // Use the current state of the elevator.

	// Correctly using the exported function name with the `checkpoint.` prefix.
	err := checkpoint.UpdataJSONOnbtnEvent(elevatorName, localElevator, checkpoint.FilenameHRAInput)
	if err != nil {
		logrus.Error("Failed to update elevator state: %v\n", err)
	}
}

func FsmHallRequestAssigner() {
	//todo, trenger dynamiske navn 
	elevatorName := "one" 

	hraInput, err := checkpoint.LoadHRAInput(checkpoint.FilenameHRAInput)
	if err != nil {
		logrus.Error("Failed to load HRAInput: %v\n", err)
		return 
	}

	jsonBytes, err := json.Marshal(hraInput)
	if err != nil {
		logrus.Error("Failed to marshal HRAInput: %v\n", err)
		return
	}


	ret, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		logrus.Error("exec.Command error: ", err)
		logrus.Error(string(ret))
		return
	}


	output := new(map[string][][2]bool) 
	err = json.Unmarshal(ret, output)
	if err != nil {
		logrus.Error("json.Unmarshal error: ", err)
		return
	}

	fmt.Println("Output:")
	for k, v := range *output {
		fmt.Printf("%6s :  %+v\n", k, v)
	}

	for floor := 0; floor < elevio.NFloors; floor++ {
        elevator.Requests[floor][elevio.BHallUp] = (*output)[elevatorName][floor][0]
        elevator.Requests[floor][elevio.BHallDown] = (*output)[elevatorName][floor][1]
        // Preserve Cab requests as is
    }
}



func FsmUpdateLocalElevatorToJSON(){
	checkpoint.UpdateLocalElevatorToJSON("one", checkpoint.FilenameHRAInput, elevator)
}













func FsmInitCyclicCounter() {
	checkpoint.InitCyclicCounter(checkpoint.FilenameCylickCounter)
}

func FsmUpdateCylickCounterButtonPressed(btnFloor int, btn elevio.Button){
	loadedCyclicCounter, err := checkpoint.LoadCyclicCounterInput(checkpoint.FilenameCylickCounter)
	if err != nil {
		logrus.Error("Feil ved lasting av CyclicCounterInput:", err)
		return
	}

	_, ok := loadedCyclicCounter.States["one"]
	if !ok{
		fmt.Println("Heisen 'one' finnes ikke i cyclicCounter.States")
	}else{
		if btn == elevio.BCab {
			loadedCyclicCounter.States["one"].CabRequests[btnFloor] += 1
		}else{

			loadedCyclicCounter.HallRequests[btnFloor][btn] += 1
		}
	}
	err = checkpoint.SaveCyclicCounterInput(loadedCyclicCounter, checkpoint.FilenameCylickCounter)
	if err != nil {
		logrus.Error("Feil ved lagring av CyclicCounterInput:", err)
		return
	}
}

func FsmUpdateCylickCounterNewFloor(){
	loadedCyclicCounter, err := checkpoint.LoadCyclicCounterInput(checkpoint.FilenameCylickCounter)
	if err != nil {
		logrus.Error("Feil ved lasting av CyclicCounterInput:", err)
		return
	}
    if state, ok := loadedCyclicCounter.States["one"]; ok {
        state.Floor += 1
        loadedCyclicCounter.States["one"] = state
    } else {
        fmt.Println("Heisen 'one' finnes ikke i cyclicCounter.States")
    }
	err = checkpoint.SaveCyclicCounterInput(loadedCyclicCounter, checkpoint.FilenameCylickCounter)
	if err != nil {
		logrus.Error("Feil ved lagring av CyclicCounterInput:", err)
		return
	}
}

/*
func UpdateCylickCounterNewbehaviour(){
	if state, ok := cyclicCounter.States["one"]; ok {
        state.Behavior += 1
        cyclicCounter.States["one"] = state
    } else {
        fmt.Println("Heisen 'one' finnes ikke i cyclicCounter.States")
    }
}


/*
ulike case er : 
-behaviour +1
-update direction +1 
*/
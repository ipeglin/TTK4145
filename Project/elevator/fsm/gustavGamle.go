package fsm

// ///AV GUSTAV TESTER OG FUCKER
//////NB HUSK Å BYTTE one med id til heis 

/*
//her er det for å oppdatere etter buttoncall 
func FsmUpdataJSONOnbtnEvent() {
	elevatorName := "one"     // Example elevator name, adjust as necessary.
	localElevator := elevator // Use the current state of the elevator.

	// Correctly using the exported function name with the `checkpoint.` prefix.
	err := checkpoint.UpdataJSONOnbtnEvent(elevatorName, localElevator, checkpoint.FilenameHRAInput)
	if err != nil {
		fmt.Printf("Failed to update elevator state: %v\n", err)
	}
}

func FsmHallRequestAssigner() {
	//todo, trenger dynamiske navn 
	elevatorName := "one" 

	hraInput, err := checkpoint.LoadHRAInput(checkpoint.FilenameHRAInput)
	if err != nil {
		fmt.Printf("Failed to load HRAInput: %v\n", err)
		return 
	}

	jsonBytes, err := json.Marshal(hraInput)
	if err != nil {
		fmt.Printf("Failed to marshal HRAInput: %v\n", err)
		return
	}


	ret, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}


	output := new(map[string][][2]bool) 
	err = json.Unmarshal(ret, output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
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
		fmt.Println("Feil ved lasting av CyclicCounterInput:", err)
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
		fmt.Println("Feil ved lagring av CyclicCounterInput:", err)
		return
	}
}

func FsmUpdateCylickCounterNewFloor(){
	loadedCyclicCounter, err := checkpoint.LoadCyclicCounterInput(checkpoint.FilenameCylickCounter)
	if err != nil {
		fmt.Println("Feil ved lasting av CyclicCounterInput:", err)
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
		fmt.Println("Feil ved lagring av CyclicCounterInput:", err)
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

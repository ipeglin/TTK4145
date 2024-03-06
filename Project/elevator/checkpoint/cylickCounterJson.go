package checkpoint

import (
	"elevator/elevio"
)
type CyclicCounterState struct {
	Behavior    int   `json:"behaviour"`
	Floor       int   `json:"floor"`
	Direction   int   `json:"direction"`
	CabRequests []int `json:"cabRequests"`
}

type CyclicCounterInput struct {
	HallRequests [][2]int                            `json:"hallRequests"`
	States       map[string]CyclicCounterState       `json:"states"`
}

func InitializeCyclicCounterInput(ElevatorName string) CyclicCounterInput {
	cyclicCounter := CyclicCounterInput{
		HallRequests: make([][2]int, elevio.NFloors),
		States:       make(map[string]CyclicCounterState), // Initialiserer map her
	}

	// Nå som States er initialisert, kan du legge til oppføringer i den
	cyclicCounter.States[ElevatorName] = CyclicCounterState{
		Behavior:    0,
		Floor:       0,
		Direction:   0,
		CabRequests: make([]int, elevio.NFloors),
	}

	return cyclicCounter
}

func updateLocalElevatorsCyclicCounterInput(cyclicCounter CyclicCounterInput, elevatorName string) CyclicCounterInput {
    state := cyclicCounter.States[elevatorName]
    state.Behavior += 1
    state.Floor += 1
    state.Direction += 1
    for i := range state.CabRequests {
        state.CabRequests[i] += 1
    }
    cyclicCounter.States[elevatorName] = state
    return cyclicCounter
}

func updateCyclicCounterWhenHallOrderIsComplete(cyclicCounter CyclicCounterInput, orderCompleteFloor int, elevatorName string) CyclicCounterInput{
	//dette er feil. må finne god logikk som finner ut hvilken av dem som skal oppdateres. 
	//trenger simen sin framgangsmåte da.
	//ser for meg det er et enkelt funksjonskall som henter knappen som skal klareres
	cyclicCounter.HallRequests[orderCompleteFloor][0] += 1
	cyclicCounter.HallRequests[orderCompleteFloor][1] += 1
	cyclicCounter = updateLocalElevatorsCyclicCounterInput(cyclicCounter, elevatorName)
	return cyclicCounter
}

func updateCyclicCounterWhenNewOrderOccurs(cyclicCounter CyclicCounterInput, hraInput HRAInput, elevatorName string,btnFloor int, btn elevio.Button)CyclicCounterInput{
    switch btn {
    case elevio.BHallUp:
		if !hraInput.HallRequests[btnFloor][0]{
        	cyclicCounter.HallRequests[btnFloor][0] += 1
		}
    case elevio.BHallDown:
		if !hraInput.HallRequests[btnFloor][1]{
			cyclicCounter.HallRequests[btnFloor][1] += 1
		}
    case elevio.BCab:
		cyclicCounter.States[elevatorName].CabRequests[btnFloor] +=1
	}
	return cyclicCounter
}


/*
func SaveCyclicCounterInput(cyclicCounter CyclicCounterInput, filename string) error {
	data, err := json.MarshalIndent(cyclicCounter, "", "  ")
	if err != nil {
		return err
	}
	osFile, err := filehandeling.LockFile(filename)
	if err != nil {
		return err

	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}


func LoadCyclicCounterInput(filename string) (CyclicCounterInput, error)  {
	osFile, err := filehandeling.LockFile(filename) // Lock the file for reading
	if err != nil {
		return CyclicCounterInput{}, err
	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	var cyclicCounter CyclicCounterInput
	data, err := os.ReadFile(filename)
	if err != nil {
		return CyclicCounterInput{}, fmt.Errorf("failed to read file: %v", err)
	}
	err = json.Unmarshal(data, &cyclicCounter)
	if err != nil {
		return CyclicCounterInput{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return cyclicCounter, nil
}



func createCyclicJSON(filename string) error {
	// Create a default HRAInput. Modify this according to your requirements.
	defaultCyclikCounter := CyclicCounterInput{
		HallRequests: make([][2]int, elevio.NFloors),
		States:       make(map[string]CyclicCounterState),
	}
	// Optionally, add default states or other initial configurations here.

	return SaveCyclicCounterInput(defaultCyclikCounter, filename)
}
*/
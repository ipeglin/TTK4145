package checkpoint

import (
	"elevator/elev"
	"elevator/elevio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

//const JSONFile = "JSONFile.json"

// const ElevatorName = "one"

// CombinedInput kombinerer HRAInput og CyclicCounterInput.
type CombinedInput struct {
	HRAInput      HRAInput
	CyclicCounter CyclicCounterInput
}

func InitializeCombinedInput(el elev.Elevator, ElevatorName string) CombinedInput {
	return CombinedInput{
		HRAInput:      initializeHRAInput(el, ElevatorName),       // Anta at denne funksjonen initialiserer HRAInput
		CyclicCounter: InitializeCyclicCounterInput(ElevatorName), // Bruker eksisterende initialiseringsfunksjon
	}
}

// SaveCombinedInput serialiserer CombinedInput til JSON og lagrer det i en fil.
func SaveCombinedInput(combinedInput CombinedInput, filename string) error {
	data, err := json.MarshalIndent(combinedInput, "", "  ")
	if err != nil {
		return fmt.Errorf("kunne ikke serialisere CombinedInput til JSON: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("kunne ikke skrive CombinedInput til fil: %v", err)
	}

	return nil
}

// LoadCombinedInput deserialiserer CombinedInput fra en JSON-fil.
func LoadCombinedInput(filename string) (CombinedInput, error) {
	var combinedInput CombinedInput

	data, err := os.ReadFile(filename)
	if err != nil {
		return combinedInput, fmt.Errorf("kunne ikke lese fil: %v", err)
	}

	err = json.Unmarshal(data, &combinedInput)
	if err != nil {
		return combinedInput, fmt.Errorf("kunne ikke deserialisere CombinedInput fra JSON: %v", err)
	}

	return combinedInput, nil
}

func UpdateJSON(el elev.Elevator, filename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(filename)
	combinedInput.HRAInput = updateHRAInput(combinedInput.HRAInput, el, elevatorName)
	combinedInput.CyclicCounter = updateLocalElevatorsCyclicCounterInput(combinedInput.CyclicCounter, elevatorName)
	SaveCombinedInput(combinedInput, filename)
}

func UpdateJSONWhenHallOrderIsComplete(el elev.Elevator, filename string, elevatorName string, orderCompleteFloor int) {
	combinedInput, _ := LoadCombinedInput(filename)
	combinedInput.HRAInput = updateHRAInputWhenHallOrderIsComplete(el, elevatorName, orderCompleteFloor)
	combinedInput.CyclicCounter = updateCyclicCounterWhenHallOrderIsComplete(combinedInput.CyclicCounter, orderCompleteFloor, elevatorName)
	SaveCombinedInput(combinedInput, filename)
}

func UpdateJSONWhenNewOrderOccurs(filename string, elevatorName string, btnFloor int, btn elevio.Button, el *elev.Elevator) {
	combinedInput, _ := LoadCombinedInput(filename)
	combinedInput.CyclicCounter = updateCyclicCounterWhenNewOrderOccurs(combinedInput.CyclicCounter, combinedInput.HRAInput, elevatorName, btnFloor, btn)
	combinedInput.HRAInput = updateHRAInputWhenNewOrderOccurs(combinedInput.HRAInput, elevatorName, btnFloor, btn, el)
	SaveCombinedInput(combinedInput, filename)
}

func JSONOrderAssigner(el *elev.Elevator, filename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(filename)

	jsonBytes, err := json.Marshal(combinedInput.HRAInput)
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

	for floor := 0; floor < elevio.NFloors; floor++ {
		el.Requests[floor][elevio.BHallUp] = (*output)[elevatorName][floor][0]
		el.Requests[floor][elevio.BHallDown] = (*output)[elevatorName][floor][1]
		// Preserve Cab requests as is
	}
}

func UpdateLocalJSON(localFilname string, incommigFilname string) {
	localCombinedInput, _ := LoadCombinedInput(localFilname)
	otherCombinedInput, _ := LoadCombinedInput(incommigFilname)

	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] > localCombinedInput.CyclicCounter.HallRequests[f][i] {
				localCombinedInput.CyclicCounter.HallRequests[f][i] = otherCombinedInput.CyclicCounter.HallRequests[f][i]
				localCombinedInput.HRAInput.HallRequests[f][i] = otherCombinedInput.HRAInput.HallRequests[f][i]
			}
		}
	}
	// trenger man forloop inne i her for alle behaviours i state?
	for i, state := range otherCombinedInput.CyclicCounter.States {
		if _, exists := localCombinedInput.CyclicCounter.States[i]; !exists {
			localCombinedInput.HRAInput.States[i] = otherCombinedInput.HRAInput.States[i]
			localCombinedInput.CyclicCounter.States[i] = state
		} else {
			if state.Behavior > localCombinedInput.CyclicCounter.States[i].Behavior {
				localCombinedInput.HRAInput.States[i] = otherCombinedInput.HRAInput.States[i]
				localCombinedInput.CyclicCounter.States[i] = state
			}
		}
	}
	SaveCombinedInput(localCombinedInput, localFilname)
}

func DeleteInactiveElevatorsFromJSON(inactiveElevatorIDs []string, localFilename string) error {
	localCombinedInput, err := LoadCombinedInput(localFilename)
	if err != nil {
		return fmt.Errorf("failed to load local combined input: %v", err)
	}

	// Convert slice of inactive elevator IDs to a map for efficient lookups
	inactiveElevatorsMap := make(map[string]struct{})
	for _, id := range inactiveElevatorIDs {
		inactiveElevatorsMap[id] = struct{}{}
	}

	// Iterate through the States in HRAInput and remove inactive elevators
	for id := range localCombinedInput.HRAInput.States {
		if _, exists := inactiveElevatorsMap[id]; exists {
			delete(localCombinedInput.HRAInput.States, id)
			delete(localCombinedInput.CyclicCounter.States, id)
		}
	}

	// Save the updated CombinedInput back to the file
	err = SaveCombinedInput(localCombinedInput, localFilename)
	if err != nil {
		return fmt.Errorf("failed to save updated combined input: %v", err)
	}

	return nil
}

func InncommingJSONHandeling(localFilname string, incommigFilname string, inncommingCombinedInput CombinedInput, inactiveElevatorIDs []string){
	SaveCombinedInput(inncommingCombinedInput, incommigFilname) 
	UpdateLocalJSON(localFilname, incommigFilname)
	DeleteInactiveElevatorsFromJSON(inactiveElevatorIDs, localFilname) 
}


func RemoveDisfunctionalElevatorFromJSON(localFilname string, elevatorName string){
	combinedInput, _ := LoadCombinedInput(localFilname)
	for id := range combinedInput.HRAInput.States {
		if (id== elevatorName) {
			delete(combinedInput.HRAInput.States, id)
			delete(combinedInput.CyclicCounter.States, id)
		}
	}
}
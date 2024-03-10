package checkpoint

import (
	"elevator/elev"
	"elevator/elevio"
	"elevator/filehandeling"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	//"strings"
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
	osFile, err := filehandeling.LockFile(filename)
	if err != nil {
		return err

	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

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
	osFile, err := filehandeling.LockFile(filename) // Lock the file for reading
	if err != nil {
		return combinedInput, err
	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

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
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		combinedInput.HRAInput = updateHRAInput(combinedInput.HRAInput, el, elevatorName)
		combinedInput.CyclicCounter = updateCyclicCounterInput(combinedInput.CyclicCounter, elevatorName)
	}
	SaveCombinedInput(combinedInput, filename)
}

func RebootJSON(el elev.Elevator, filename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(filename)
	combinedInput.HRAInput = rebootHRAInput(combinedInput.HRAInput, el, elevatorName)
	combinedInput.CyclicCounter = updateCyclicCounterInput(combinedInput.CyclicCounter, elevatorName)
	SaveCombinedInput(combinedInput, filename)
}

func UpdateJSONWhenHallOrderIsComplete(el elev.Elevator, filename string, elevatorName string, btn_floor int, btn_type elevio.Button) {
	combinedInput, _ := LoadCombinedInput(filename)
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		combinedInput.HRAInput = updateHRAInputWhenOrderIsComplete(combinedInput.HRAInput, el, elevatorName, btn_floor, btn_type)
		combinedInput.CyclicCounter = updateCyclicCounterWhenOrderIsComplete(combinedInput.CyclicCounter, elevatorName, btn_floor, btn_type)
	}
	SaveCombinedInput(combinedInput, filename)
}

func UpdateJSONWhenNewOrderOccurs(filename string, elevatorName string, btnFloor int, btn elevio.Button, el *elev.Elevator) {
	combinedInput, _ := LoadCombinedInput(filename)
	
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		combinedInput.CyclicCounter = updateCyclicCounterWhenNewOrderOccurs(combinedInput.CyclicCounter, combinedInput.HRAInput, elevatorName, btnFloor, btn)
		combinedInput.HRAInput = updateHRAInputWhenNewOrderOccurs(combinedInput.HRAInput, elevatorName, btnFloor, btn, el)
	}
	SaveCombinedInput(combinedInput, filename)
}

func JSONOrderAssigner(el *elev.Elevator, filename string, elevatorName string) {
	combinedInput, err := LoadCombinedInput(filename)
	if err != nil {
		fmt.Printf("Failed to load combined input: %v\n", err)
		return
	}

	// Check if HRAInput.States is not empty
	if len(combinedInput.HRAInput.States) > 0 {
		jsonBytes, err := json.Marshal(combinedInput.HRAInput)
		if err != nil {
			fmt.Printf("Failed to marshal HRAInput: %v\n", err)
			return
		}

		ret, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
		if err != nil {
			fmt.Printf("exec.Command error: %v\nOutput: %s\n", err, string(ret))
			return
		}

		output := make(map[string][][2]bool) // Changed from using new to make for clarity
		if err := json.Unmarshal(ret, &output); err != nil {
			fmt.Printf("json.Unmarshal error: %v\n", err)
			return
		}

		for floor := 0; floor < elevio.NFloors; floor++ {
			if orders, ok := output[elevatorName]; ok && floor < len(orders) {
				el.Requests[floor][elevio.BHallUp] = orders[floor][0]
				el.Requests[floor][elevio.BHallDown] = orders[floor][1]
			}
		}
	} else {
		fmt.Println("HRAInput.States is empty, skipping order assignment")
	}
}

/*
func UpdateLocalJSON(localFilname string, otherCombinedInput CombinedInput, incomingFilename string) {
	localCombinedInput, _ := LoadCombinedInput(localFilname)
	//otherCombinedInput, _ := LoadCombinedInput(incomingFilename)

	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] > localCombinedInput.CyclicCounter.HallRequests[f][i] {
				localCombinedInput.CyclicCounter.HallRequests[f][i] = otherCombinedInput.CyclicCounter.HallRequests[f][i]
				localCombinedInput.HRAInput.HallRequests[f][i] = otherCombinedInput.HRAInput.HallRequests[f][i]
			}
		}
	}
	incommigElevatorName := strings.TrimSuffix(incomingFilename, ".json")
	if _, exists := otherCombinedInput.CyclicCounter.States[incommigElevatorName]; exists{
		if _, exists := localCombinedInput.CyclicCounter.States[incommigElevatorName]; !exists {
			localCombinedInput.HRAInput.States[incommigElevatorName] = otherCombinedInput.HRAInput.States[incommigElevatorName]
			localCombinedInput.CyclicCounter.States[incommigElevatorName] = otherCombinedInput.CyclicCounter.States[incommigElevatorName]
		} else {
			if otherCombinedInput.CyclicCounter.States[incommigElevatorName] > localCombinedInput.CyclicCounter.States[incommigElevatorName] {
				localCombinedInput.HRAInput.States[incommigElevatorName] = otherCombinedInput.HRAInput.States[incommigElevatorName]
				localCombinedInput.CyclicCounter.States[incommigElevatorName] = otherCombinedInput.CyclicCounter.States[incommigElevatorName]

			}
		}
	}

	//legg kunn til deres local elevator 
	/*
	for i, state := range otherCombinedInput.HRAInput.States {
		if _, exists := localCombinedInput.CyclicCounter.States[i]; !exists {
			localCombinedInput.HRAInput.States[i] = state
			localCombinedInput.CyclicCounter.States[i] = otherCombinedInput.CyclicCounter.States[i]
		} else {
			if otherCombinedInput.CyclicCounter.States[i] > localCombinedInput.CyclicCounter.States[i] {
				localCombinedInput.HRAInput.States[i] = state
				localCombinedInput.CyclicCounter.States[i] = otherCombinedInput.CyclicCounter.States[i]
			}
		}
	}


	SaveCombinedInput(localCombinedInput, localFilname)
}
*/
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


func InncommingJSONHandeling(localFilname string, otherCombinedInput CombinedInput, incomingElevatorName string) {
	localCombinedInput, _ := LoadCombinedInput(localFilname)
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
		print("slett den jævla heisen")
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; exists {
			delete(localCombinedInput.HRAInput.States, incomingElevatorName)
			delete(localCombinedInput.CyclicCounter.States, incomingElevatorName)
		}
	}
	SaveCombinedInput(localCombinedInput, localFilname)
}


func RemoveDysfunctionalElevatorFromJSON(localFilname string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(localFilname)
	for id := range combinedInput.HRAInput.States {
		if id == elevatorName {
			delete(combinedInput.HRAInput.States, id)
			delete(combinedInput.CyclicCounter.States, id)
		}
	}
	SaveCombinedInput(combinedInput, localFilname)
}

/*
func DysfunctionalElevatorDetection(incomingFilename string, incomingCombinedInput CombinedInput) []string {
	incomingElevatorName := strings.TrimSuffix(incomingFilename, ".json")
	var inactiveElevatorIDs []string
	if _, exists := incomingCombinedInput.HRAInput.States[incomingElevatorName]; !exists {
		println(incomingElevatorName)
		inactiveElevatorIDs = append(inactiveElevatorIDs, incomingElevatorName)
	}
	return inactiveElevatorIDs
}
*/

// Antagelser om strukturer og hjelpefunksjoner fra tidligere eksempel ...
// IsValidBehavior sjekker om oppgitt atferd er gyldig
func IsValidBehavior(behavior string) bool {
	switch behavior {
	case "idle", "moving", "doorOpen":
		return true
	default:
		return false
	}
}

// IsValidDirection sjekker om oppgitt retning er gyldig
func IsValidDirection(direction string) bool {
	switch direction {
	case "up", "down", "stop":
		return true
	default:
		return false
	}
}

// IncomingDataIsCorrupt sjekker om inngående data er korrupt
func IncomingDataIsCorrupt(incomingCombinedInput CombinedInput) bool {
	incomingHRAInput := incomingCombinedInput.HRAInput
	if len(incomingHRAInput.HallRequests) != elevio.NFloors {
		return true
	}
	for _, state := range incomingHRAInput.States {
		if !IsValidBehavior(state.Behavior) || !IsValidDirection(state.Direction) {
			return true // Data er korrupt basert på ugyldig Behavior eller Direction
		}

		// Sjekk om CabRequests har riktig lengde og inneholder boolske verdier
		if len(state.CabRequests) != elevio.NFloors {
			return true // Data er korrupt basert på lengde
		}
	}
	return false // Data er gyldig
}

func JSONsetAllLights(localFilname string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(localFilname)
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		for floor := 0; floor < elevio.NFloors; floor++ {
			elevio.RequestButtonLight(floor, elevio.BHallUp, combinedInput.HRAInput.HallRequests[floor][0])
			elevio.RequestButtonLight(floor, elevio.BHallDown, combinedInput.HRAInput.HallRequests[floor][1])
		}
	}
}

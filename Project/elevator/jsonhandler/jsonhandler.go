package jsonhandler

import (
	ccounter "elevator/cycliccounter"
	"elevator/elev"
	"elevator/elevio"
	"elevator/filehandler"
	"elevator/hra"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	// "strings"

	"github.com/sirupsen/logrus"
)

// CombinedInput kombinerer HRAInput og CyclicCounterInput.
type CombinedInput struct {
	HRAInput      hra.HRAInput
	CyclicCounter ccounter.CyclicCounterInput
}

func InitializeCombinedInput(el elev.Elevator, ElevatorName string) CombinedInput {
	return CombinedInput{
		HRAInput:      hra.InitializeHRAInput(el, ElevatorName),       // Anta at denne funksjonen initialiserer HRAInput
		CyclicCounter: ccounter.InitializeCyclicCounterInput(ElevatorName), // Bruker eksisterende initialiseringsfunksjon
	}
}

// SaveCombinedInput serialiserer CombinedInput til JSON og lagrer det i en fil.
func SaveCombinedInput(combinedInput CombinedInput, filename string) error {
	osFile, err := filehandler.LockFile(filename)
	if err != nil {
		return err

	}
	defer filehandler.UnlockFile(osFile) // Ensure file is unlocked after reading

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
	osFile, err := filehandler.LockFile(filename) // Lock the file for reading
	if err != nil {
		return combinedInput, err
	}
	defer filehandler.UnlockFile(osFile) // Ensure file is unlocked after reading

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
		combinedInput.HRAInput = hra.UpdateHRAInput(combinedInput.HRAInput, el, elevatorName)
		combinedInput.CyclicCounter = ccounter.UpdateInput(combinedInput.CyclicCounter, elevatorName)
	}
	SaveCombinedInput(combinedInput, filename)
}

// This was RebootJSON
func UpdateJSONOnReboot(el elev.Elevator, filename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(filename)
	combinedInput.HRAInput = hra.RebootHRAInput(combinedInput.HRAInput, el, elevatorName)
	combinedInput.CyclicCounter = ccounter.UpdateInput(combinedInput.CyclicCounter, elevatorName)
	SaveCombinedInput(combinedInput, filename)
}

func UpdateJSONOnCompletedHallOrder(el elev.Elevator, filename string, elevatorName string, btn_floor int, btn_type elevio.Button) {
	combinedInput, _ := LoadCombinedInput(filename)
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		combinedInput.HRAInput = hra.UpdateHRAInputOnCompletedOrder(combinedInput.HRAInput, el, elevatorName, btn_floor, btn_type)
		combinedInput.CyclicCounter = ccounter.UpdateOnCompletedOrder(combinedInput.CyclicCounter, elevatorName, btn_floor, btn_type)
	}
	SaveCombinedInput(combinedInput, filename)
}

func UpdateJSONOnNewOrder(filename string, elevatorName string, btnFloor int, btn elevio.Button) {
	combinedInput, _ := LoadCombinedInput(filename)
	//ønsker vi ikke legge til nye ordere/ ta ordere mens vi er offline? 
	//hvis vi øssker, fjern denne if setningen. 
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		combinedInput.CyclicCounter = ccounter.UpdateOnNewOrder(combinedInput.CyclicCounter, combinedInput.HRAInput, elevatorName, btnFloor, btn)
		combinedInput.HRAInput = hra.UpdateHRAInputWhenNewOrderOccurs(combinedInput.HRAInput, elevatorName, btnFloor, btn)
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
		logrus.Debug("HRAInput.States is empty, skipping order assignment")
	}
}
/*
func HandleIncomingJSON(el elev.Elevator,localFilename string, localElevatorName string, otherCombinedInput CombinedInput, incomingElevatorName string) {
	localCombinedInput, _ := LoadCombinedInput(localFilename)
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] > localCombinedInput.CyclicCounter.HallRequests[f][i] {
				localCombinedInput.CyclicCounter.HallRequests[f][i] = otherCombinedInput.CyclicCounter.HallRequests[f][i]
				localCombinedInput.HRAInput.HallRequests[f][i] = otherCombinedInput.HRAInput.HallRequests[f][i]
			}
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] == localCombinedInput.CyclicCounter.HallRequests[f][i] {
				if localCombinedInput.HRAInput.HallRequests[f][i] != otherCombinedInput.HRAInput.HallRequests[f][i] {
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
	} else {
		if _, exists := localCombinedInput.HRAInput.States[incomingElevatorName]; exists {
			delete(localCombinedInput.HRAInput.States, incomingElevatorName)
			delete(localCombinedInput.CyclicCounter.States, incomingElevatorName)
		}
	}
	if _, exists := otherCombinedInput.CyclicCounter.States[localElevatorName]; exists {
		if otherCombinedInput.CyclicCounter.States[localElevatorName] > localCombinedInput.CyclicCounter.States[localElevatorName] {
			localCombinedInput.CyclicCounter.States[localElevatorName] = otherCombinedInput.CyclicCounter.States[localElevatorName] + 1
		}
	}
	SaveCombinedInput(localCombinedInput, localFilename)
	allValuesGreater := true
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if otherCombinedInput.CyclicCounter.HallRequests[f][i] < localCombinedInput.CyclicCounter.HallRequests[f][i] {
				allValuesGreater = false
				break
			}
		}
	}
	if allValuesGreater {
		// Execute further actions here
		JSONsetAllLights(localFilename, localElevatorName)
		JSONOrderAssigner(& el, localFilename, localElevatorName)
		//fsm.MoveOnActiveOrders(localFilename, localElevatorName) // TODO: Only have one version
	}
}
*/
func RemoveDysfunctionalElevatorFromJSON(localFilename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(localFilename)
	for id := range combinedInput.HRAInput.States {
		if id == elevatorName {
			delete(combinedInput.HRAInput.States, id)
			delete(combinedInput.CyclicCounter.States, id)
		}
	}
	SaveCombinedInput(combinedInput, localFilename)
}

// denne brukes en gang i main. kan vi gjøre den over komatibel
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
			//ønsker ikke fjerne cylick counter
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

// SIMEN HEVDER DETTE ER DOBBELT OPP AV FUNKSJONER
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

func JSONsetAllLights(localFilename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(localFilename)
	//if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
	for floor := 0; floor < elevio.NFloors; floor++ {
		elevio.RequestButtonLight(floor, elevio.BHallUp, combinedInput.HRAInput.HallRequests[floor][0])
		elevio.RequestButtonLight(floor, elevio.BHallDown, combinedInput.HRAInput.HallRequests[floor][1])
	
	}
}

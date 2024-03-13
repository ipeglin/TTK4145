package jsonhandler

import (
	ccounter "elevator/cycliccounter"
	"elevator/elev"
	"elevator/elevio"
	"elevator/hra"
	"encoding/json"
	"filehandler"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// TODO: Change to ElevatorState or similar
// CombinedInput kombinerer HRAInput og CyclicCounterInput.
type CombinedInput struct {
	HRAInput      hra.HRAInput
	CyclicCounter ccounter.CyclicCounterInput
}

func InitializeCombinedInput(e elev.Elevator, elevatorName string) CombinedInput {
	return CombinedInput{
		HRAInput:      hra.InitializeHRAInput(e, elevatorName),
		CyclicCounter: ccounter.InitializeCyclicCounterInput(elevatorName),
	}
}

// SaveCombinedInput serialiserer CombinedInput til JSON og lagrer det i en fil.
func SaveCombinedInput(combinedInput CombinedInput, filename string) error {
	data, err := json.MarshalIndent(combinedInput, "", "  ")
	if err != nil {
		return fmt.Errorf("kunne ikke serialisere CombinedInput til JSON: %v", err)
	}

	filehandler.WriteToFile(data, filename)
	return nil
}

// LoadCombinedInput deserialiserer CombinedInput fra en JSON-fil.
func LoadCombinedInput(filename string) (CombinedInput, error) {
	var combinedInput CombinedInput

	data, err := filehandler.ReadFromFile(filename)
	if err != nil {
		logrus.Error("Failed to read from file: ", err)
		return combinedInput, fmt.Errorf("failed to read from file %v", err)
	}

	err = json.Unmarshal(data, &combinedInput)
	if err != nil {
		return combinedInput, fmt.Errorf("kunne ikke deserialisere CombinedInput fra JSON: %v", err)
	}

	return combinedInput, nil
}

func UpdateJSON(e elev.Elevator, filename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(filename)
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		combinedInput.HRAInput = hra.UpdateHRAInput(combinedInput.HRAInput, e, elevatorName)
		combinedInput.CyclicCounter = ccounter.IncrementOnInput(combinedInput.CyclicCounter, elevatorName)
	}
	SaveCombinedInput(combinedInput, filename)
}

// This was RebootJSON
func UpdateJSONOnReboot(e elev.Elevator, filename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(filename)
	combinedInput.HRAInput = hra.RebootHRAInput(combinedInput.HRAInput, e, elevatorName)
	combinedInput.CyclicCounter = ccounter.IncrementOnInput(combinedInput.CyclicCounter, elevatorName)
	SaveCombinedInput(combinedInput, filename)
}

func UpdateJSONOnCompletedHallOrder(e elev.Elevator, filename string, elevatorName string, btn_floor int, btn_type elevio.Button) {
	combinedInput, _ := LoadCombinedInput(filename)
	if _, exists := combinedInput.HRAInput.States[elevatorName]; exists {
		combinedInput.HRAInput = hra.UpdateHRAInputOnCompletedOrder(combinedInput.HRAInput, e, elevatorName, btn_floor, btn_type)
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
		combinedInput.HRAInput = hra.UpdateHRAInputOnNewOrder(combinedInput.HRAInput, elevatorName, btnFloor, btn)
	}
	SaveCombinedInput(combinedInput, filename)
}

// TODO : mener denne kan bare bli en fsm func
// TODO: Changed it from refrence to pass-by-value. Is it very ugly now?
func JSONOrderAssigner(e elev.Elevator, filename string, elevatorName string) elev.Elevator {
	combinedInput, err := LoadCombinedInput(filename)
	if err != nil {
		fmt.Printf("Failed to load combined input: %v\n", err)
		return e
	}

	// Check if HRAInput.States is not empty
	if len(combinedInput.HRAInput.States) > 0 {
		jsonBytes, err := json.Marshal(combinedInput.HRAInput)
		if err != nil {
			fmt.Printf("Failed to marshal HRAInput: %v\n", err)
			return e
		}

		ret, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
		if err != nil {
			fmt.Printf("exec.Command error: %v\nOutput: %s\n", err, string(ret))
			return e
		}

		output := make(map[string][][2]bool) // Changed from using new to make for clarity
		if err := json.Unmarshal(ret, &output); err != nil {
			fmt.Printf("json.Unmarshal error: %v\n", err)
			return e
		}

		for floor := 0; floor < elevio.NFloors; floor++ {
			if orders, ok := output[elevatorName]; ok && floor < len(orders) {
				e.Requests[floor][elevio.BHallUp] = orders[floor][0]
				e.Requests[floor][elevio.BHallDown] = orders[floor][1]
			}
		}
		return e
	} else {
		logrus.Debug("HRAInput.States is empty, skipping order assignment")
		return e
	}
}

func RemoveDysfunctionalElevatorFromJSON(localStateFilename string, elevatorName string) {
	combinedInput, _ := LoadCombinedInput(localStateFilename)
	for id := range combinedInput.HRAInput.States {
		if id == elevatorName {
			delete(combinedInput.HRAInput.States, id)
			delete(combinedInput.CyclicCounter.States, id)
		}
	}
	SaveCombinedInput(combinedInput, localStateFilename)
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

func IsValidBehavior(behavior string) bool {
	switch behavior {
	case "idle", "moving", "doorOpen":
		return true
	default:
		return false
	}
}

func IsValidDirection(direction string) bool {
	switch direction {
	case "up", "down", "stop":
		return true
	default:
		return false
	}
}

func IncomingDataIsCorrupt(incomingCombinedInput CombinedInput) bool {
	incomingHRAInput := incomingCombinedInput.HRAInput
	if len(incomingHRAInput.HallRequests) != elevio.NFloors {
		return true
	}
	for _, state := range incomingHRAInput.States {
		if !IsValidBehavior(state.Behavior) || !IsValidDirection(state.Direction) {
			return true
		}

		if len(state.CabRequests) != elevio.NFloors {
			return true
		}
	}
	return false
}

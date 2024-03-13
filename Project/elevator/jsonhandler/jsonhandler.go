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

// TElevState kombinerer HRAInput og CyclicCounterInput.
type TElevState struct {
	HRAInput      hra.HRAInput
	CyclicCounter ccounter.CyclicCounterInput
}

// TODO: Change this from Initizalied to make/create
func InitialiseState(e elev.Elevator, elevatorName string) TElevState {
	return TElevState{
		HRAInput:      hra.InitializeHRAInput(e, elevatorName),
		CyclicCounter: ccounter.InitializeCyclicCounterInput(elevatorName),
	}
}

// SaveState serialiserer state til JSON og lagrer det i en fil.
func SaveState(state TElevState, filename string) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialise state to JSON: %v", err)
	}

	filehandler.WriteToFile(data, filename)
	return nil
}

// LoadState deserialiserer state fra en JSON-fil.
func LoadState(filename string) (TElevState, error) {
	var state TElevState

	data, err := filehandler.ReadFromFile(filename)
	if err != nil {
		logrus.Error("Failed to read from file: ", err)
		return state, fmt.Errorf("failed to read from file %v", err)
	}

	err = json.Unmarshal(data, &state)
	if err != nil {
		return state, fmt.Errorf("failed to deserialise state from JSON: %v", err)
	}

	return state, nil
}

func UpdateJSON(e elev.Elevator, filename string, elevatorName string) {
	state, _ := LoadState(filename)
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.HRAInput = hra.UpdateHRAInput(state.HRAInput, e, elevatorName)
		state.CyclicCounter = ccounter.IncrementOnInput(state.CyclicCounter, elevatorName)
	}
	SaveState(state, filename)
}

// This was RebootJSON
func UpdateJSONOnReboot(e elev.Elevator, filename string, elevatorName string) {
	state, _ := LoadState(filename)
	state.HRAInput = hra.RebootHRAInput(state.HRAInput, e, elevatorName)
	state.CyclicCounter = ccounter.IncrementOnInput(state.CyclicCounter, elevatorName)
	SaveState(state, filename)
}

func UpdateJSONOnCompletedHallOrder(e elev.Elevator, filename string, elevatorName string, btn_floor int, btn_type elevio.Button) {
	state, _ := LoadState(filename)
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.HRAInput = hra.UpdateHRAInputOnCompletedOrder(state.HRAInput, e, elevatorName, btn_floor, btn_type)
		state.CyclicCounter = ccounter.UpdateOnCompletedOrder(state.CyclicCounter, elevatorName, btn_floor, btn_type)
	}
	SaveState(state, filename)
}

func UpdateJSONOnNewOrder(filename string, elevatorName string, btnFloor int, btn elevio.Button) {
	state, _ := LoadState(filename)
	//ønsker vi ikke legge til nye ordere/ ta ordere mens vi er offline?
	//hvis vi øssker, fjern denne if setningen.
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.CyclicCounter = ccounter.UpdateOnNewOrder(state.CyclicCounter, state.HRAInput, elevatorName, btnFloor, btn)
		state.HRAInput = hra.UpdateHRAInputOnNewOrder(state.HRAInput, elevatorName, btnFloor, btn)
	}
	SaveState(state, filename)
}

// TODO : mener denne kan bare bli en fsm func
// TODO: Changed it from refrence to pass-by-value. Is it very ugly now?
func JSONOrderAssigner(e elev.Elevator, filename string, elevatorName string) elev.Elevator {
	state, err := LoadState(filename)
	if err != nil {
		fmt.Printf("Failed to load combined input: %v\n", err)
		return e
	}

	// Check if HRAInput.States is not empty
	if len(state.HRAInput.States) > 0 {
		jsonBytes, err := json.Marshal(state.HRAInput)
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

// denne brukes en gang i main. kan vi gjøre den over komatibel
func RemoveElevatorsFromJSON(elevatorIDs []string, filename string) error {
	state, err := LoadState(filename)
	if err != nil {
		return fmt.Errorf("failed to load local combined input: %v", err)
	}

	// Convert slice of inactive elevator IDs to a map for efficient lookups
	inactiveElevatorsMap := make(map[string]struct{})
	for _, id := range elevatorIDs {
		inactiveElevatorsMap[id] = struct{}{}
	}

	// Iterate through the States in HRAInput and remove inactive elevators
	for id := range state.HRAInput.States {
		if _, exists := inactiveElevatorsMap[id]; exists {
			delete(state.HRAInput.States, id)
			//ønsker ikke fjerne cylick counter
			delete(state.CyclicCounter.States, id)
		}
	}

	// Save the updated state back to the file
	//TODO: Got error check for this saveState but not for anyone else
	err = SaveState(state, filename)
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

func IncomingDataIsCorrupt(externalState TElevState) bool {
	incomingHRAInput := externalState.HRAInput
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

package jsonhandler

import (
	"elevator/counter"
	"elevator/elev"
	"elevator/elevio"
	"elevator/hra"
	"encoding/json"
	"filehandler"
	"fmt"
	"network/local"
	"os/exec"

	"github.com/sirupsen/logrus"
)

var StateFile string

// ElevatorState kombinerer HRAInput og Counter.
type ElevatorState struct {
	HRAInput hra.HRAInput
	Counter  counter.Counter
}

func init() {
	ip, err := local.GetIP()
	if err != nil {
		logrus.Fatal("Could not resolve IP address")
	}

	StateFile = ip + ".json"
}

func InitialiseState(e elev.Elevator, elevatorName string) ElevatorState {
	return ElevatorState{
		HRAInput: hra.InitialiseHRAInput(e, elevatorName),
		Counter:  counter.InitialiseCounter(elevatorName),
	}
}

// SaveState serialiserer state til JSON og lagrer det i en fil.
func SaveState(state ElevatorState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialise state to JSON: %v", err)
	}

	filehandler.WriteToFile(data, StateFile)
	return nil
}

// LoadState deserialiserer state fra en JSON-fil.
func LoadState() (ElevatorState, error) {
	var state ElevatorState

	data, err := filehandler.ReadFromFile(StateFile)
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

func UpdateJSON(e elev.Elevator, elevatorName string) {
	state, _ := LoadState()
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.HRAInput = hra.UpdateHRAInput(state.HRAInput, e, elevatorName)
		state.Counter = counter.IncrementOnInput(state.Counter, elevatorName)
	}
	SaveState(state)
}

// This was RebootJSON
func UpdateJSONOnReboot(e elev.Elevator, elevatorName string) {
	state, _ := LoadState()
	state.HRAInput = hra.RebootHRAInput(state.HRAInput, e, elevatorName)
	state.Counter = counter.IncrementOnInput(state.Counter, elevatorName)
	SaveState(state)
}

func UpdateJSONOnCompletedHallOrder(e elev.Elevator, elevatorName string, btn_floor int, btn_type elevio.Button) {
	state, _ := LoadState()
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.HRAInput = hra.UpdateHRAInputOnCompletedOrder(state.HRAInput, e, elevatorName, btn_floor, btn_type)
		state.Counter = counter.UpdateOnCompletedOrder(state.Counter, elevatorName, btn_floor, btn_type)
	}
	SaveState(state)
}

func UpdateJSONOnNewOrder(elevatorName string, btnFloor int, btn elevio.Button) {
	state, _ := LoadState()
	//ønsker vi ikke legge til nye ordere/ ta ordere mens vi er offline?
	//hvis vi øssker, fjern denne if setningen.
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.Counter = counter.UpdateOnNewOrder(state.Counter, state.HRAInput, elevatorName, btnFloor, btn)
		state.HRAInput = hra.UpdateHRAInputOnNewOrder(state.HRAInput, elevatorName, btnFloor, btn)
	}
	SaveState(state)
}

// TODO : mener denne kan bare bli en fsm func
func JSONOrderAssigner(e *elev.Elevator, elevatorName string) {
	state, err := LoadState()
	if err != nil {
		fmt.Printf("Failed to load combined input: %v\n", err)
		return
	}

	// Check if HRAInput.States is not empty
	if len(state.HRAInput.States) > 0 {
		jsonBytes, err := json.Marshal(state.HRAInput)
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
				e.Requests[floor][elevio.BHallUp] = orders[floor][0]
				e.Requests[floor][elevio.BHallDown] = orders[floor][1]
			}
		}
	} else {
		logrus.Debug("HRAInput.States is empty, skipping order assignment")
	}
}

// denne brukes en gang i main. kan vi gjøre den over komatibel
func RemoveElevatorsFromJSON(elevatorIDs []string) error {
	state, err := LoadState()
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
			delete(state.Counter.States, id)
		}
	}

	// Save the updated state back to the file
	//TODO: Got error check for this saveState but not for anyone else
	err = SaveState(state)
	if err != nil {
		return fmt.Errorf("failed to save updated combined input: %v", err)
	}

	return nil
}

func IsStateCorrupted(state ElevatorState) bool {
	input := state.HRAInput

	if len(input.HallRequests) != elevio.NFloors {
		return true
	}

	for _, state := range input.States {
		if !isValidBehavior(state.Behavior) || !isValidDirection(state.Direction) {
			return true
		}

		if len(state.CabRequests) != elevio.NFloors {
			return true
		}
	}

	return false
}

func isValidBehavior(behavior string) bool {
	switch behavior {
	case "idle", "moving", "doorOpen":
		return true
	default:
		return false
	}
}

func isValidDirection(direction string) bool {
	switch direction {
	case "up", "down", "stop":
		return true
	default:
		return false
	}
}

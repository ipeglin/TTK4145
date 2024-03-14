package statehandler

import (
	"elevator/counter"
	"elevator/elev"
	"elevator/elevio"
	"elevator/hra"
	"encoding/json"
	"filehandler"
	"fmt"
	"network/local"

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
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.Counter = counter.UpdateOnNewOrder(state.Counter, state.HRAInput, elevatorName, btnFloor, btn)
		state.HRAInput = hra.UpdateHRAInputOnNewOrder(state.HRAInput, elevatorName, btnFloor, btn)
	}
	SaveState(state)
}

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

func HandleIncomingJSON(localElevatorName string, externalState ElevatorState, incomingElevatorName string) {
	localState, _ := LoadState()
	localState = mergeWithIncomingHallRequests(localState, externalState)
	if _, exists := externalState.HRAInput.States[incomingElevatorName]; exists {
		localState.HRAInput.States[incomingElevatorName] = externalState.HRAInput.States[incomingElevatorName]
		localState.Counter.States[incomingElevatorName] = externalState.Counter.States[incomingElevatorName]
			if externalState.Counter.States[localElevatorName] > localState.Counter.States[localElevatorName] {
				localState.Counter.States[localElevatorName] = externalState.Counter.States[localElevatorName] + 1
			}
	} else {
		delete(localState.HRAInput.States, incomingElevatorName)
		delete(localState.Counter.States, incomingElevatorName)
		//RemoveElevatorsFromJSON([]string{incomingElevatorName})
	}
	SaveState(localState)
}
func mergeWithIncomingHallRequests(localState, externalState ElevatorState) ElevatorState {
	for f := 0; f < elevio.NFloors; f++ {
		for i := 0; i < 2; i++ {
			if externalState.Counter.HallRequests[f][i] > localState.Counter.HallRequests[f][i] {
				localState.Counter.HallRequests[f][i] = externalState.Counter.HallRequests[f][i]
				localState.HRAInput.HallRequests[f][i] = externalState.HRAInput.HallRequests[f][i]
			}
			if externalState.Counter.HallRequests[f][i] == localState.Counter.HallRequests[f][i] {
				if localState.HRAInput.HallRequests[f][i] != externalState.HRAInput.HallRequests[f][i] {
					localState.HRAInput.HallRequests[f][i] = false
				}
			}
		}
	}
	return localState
}
// TODO: Gustav shceck if this is neccesary
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

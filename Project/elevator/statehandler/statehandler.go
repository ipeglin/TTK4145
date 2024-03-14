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

func SaveState(state ElevatorState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialise state to JSON: %v", err)
	}

	filehandler.WriteToFile(data, StateFile)
	return nil
}

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

func UpdateState(e elev.Elevator, elevatorName string) {
	state, _ := LoadState()
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.HRAInput = hra.UpdateHRAInput(state.HRAInput, e, elevatorName)
		state.Counter = counter.IncrementOnInput(state.Counter, elevatorName)
	}
	SaveState(state)
}

func UpdateStateOnReboot(e elev.Elevator, elevatorName string) {
	state, _ := LoadState()
	state.HRAInput = hra.RebootHRAInput(state.HRAInput, e, elevatorName)
	state.Counter = counter.IncrementOnInput(state.Counter, elevatorName)
	SaveState(state)
}

func UpdateStateOnCompletedHallOrder(e elev.Elevator, elevatorName string, btn_floor int, btn_type elevio.Button) {
	state, _ := LoadState()
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.HRAInput = hra.UpdateHRAInputOnCompletedOrder(state.HRAInput, e, elevatorName, btn_floor, btn_type)
		state.Counter = counter.UpdateOnCompletedOrder(state.Counter, elevatorName, btn_floor, btn_type)
	}
	SaveState(state)
}

func UpdateStateOnNewOrder(elevatorName string, btnFloor int, btn elevio.Button) {
	state, _ := LoadState()
	if _, exists := state.HRAInput.States[elevatorName]; exists {
		state.Counter = counter.UpdateOnNewOrder(state.Counter, state.HRAInput, elevatorName, btnFloor, btn)
		state.HRAInput = hra.UpdateHRAInputOnNewOrder(state.HRAInput, elevatorName, btnFloor, btn)
	}
	SaveState(state)
}

func RemoveElevatorsFromState(elevatorIDs []string) error {
	state, err := LoadState()
	if err != nil {
		return fmt.Errorf("failed to load local combined input: %v", err)
	}

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
	//TODO: Got error check for this saveState but not for anyone else
	err = SaveState(state)
	if err != nil {
		return fmt.Errorf("failed to save updated combined input: %v", err)
	}

	return nil
}

func HandleIncomingSate(localElevatorName string, externalState ElevatorState, incomingElevatorName string) {
	localState, _ := LoadState()
	mergeWithIncomingHallRequests(localState, externalState)
	if _, exists := externalState.HRAInput.States[incomingElevatorName]; exists {
		mergeWithIncomigStates(localState, localElevatorName, externalState, incomingElevatorName)
	} else {
		RemoveElevatorsFromState([]string{incomingElevatorName})
	}
}

func mergeWithIncomingHallRequests(localState ElevatorState, externalState ElevatorState) {
	for f := 0; f < elevio.NFloors; f++ {
		for btn := elevio.BHallUp; btn < elevio.BCab; btn++ {
			if externalState.Counter.HallRequests[f][btn] > localState.Counter.HallRequests[f][btn] {
				localState.Counter.HallRequests[f][btn] = externalState.Counter.HallRequests[f][btn]
				localState.HRAInput.HallRequests[f][btn] = externalState.HRAInput.HallRequests[f][btn]
			}
			if externalState.Counter.HallRequests[f][btn] == localState.Counter.HallRequests[f][btn] {
				if localState.HRAInput.HallRequests[f][btn] != externalState.HRAInput.HallRequests[f][btn] {
					localState.HRAInput.HallRequests[f][btn] = false
				}
			}
		}
	}
	SaveState(localState)
}

func mergeWithIncomigStates(localState ElevatorState, localElevatorName string, externalState ElevatorState, incomingElevatorName string) {
	localState.HRAInput.States[incomingElevatorName] = externalState.HRAInput.States[incomingElevatorName]
	localState.Counter.States[incomingElevatorName] = externalState.Counter.States[incomingElevatorName]

	if externalState.Counter.States[localElevatorName] > localState.Counter.States[localElevatorName] {
		localState.Counter.States[localElevatorName] = externalState.Counter.States[localElevatorName] + 1
	}
	SaveState(localState)
}

func IsOnlyNodeOnline(localElevatorName string) bool {
	currentState, _ := LoadState()
	if len(currentState.HRAInput.States) == 1 {
		if _, exists := currentState.HRAInput.States[localElevatorName]; exists {
			return true
		}
	}
	return false
}

// TODO: Gustav shceck if this is neccesary
//Checked, we dont, but is it harmfull?
//it makes our code more robust, but if we feel we dont need?
/*
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
*/

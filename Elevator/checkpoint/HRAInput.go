package checkpoint

import (
	"encoding/json"
	"fmt"
	"heislab/Elevator/elev"
	"heislab/Elevator/elevio"
	"heislab/Elevator/filehandeling"
	"os"
)

const FilenameHRAInput = "elevHRAInput.JSON"

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func UpdataJSONOnbtnEvent(elevatorName string, localElevator elev.Elevator, filename string) error {
	hraInput, err := LoadHRAInput(filename)
	if err != nil {
		// If loading fails, try to initialize the HRAInput file.
		initErr := initializeHRAInput(filename)

		if initErr != nil {
			// If initialization also fails, return this new error.
			return fmt.Errorf("failed to initialize HRAInput JSON: %v", initErr)
		}
		// After initialization, attempt to load the newly created file again.
		hraInput, err = LoadHRAInput(filename)
		if err != nil {
			// If loading still fails, return this error.
			return fmt.Errorf("failed to load HRAInput from JSON after initialization: %v", err)
		}
	}
	// Convert the local elevator's state
	behavior, direction, cabRequests := convertLocalElevatorState(localElevator)

	// Update only the specified elevator's state in HRAInput
	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       localElevator.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}

	// WE COULD CHANGE THIS TO ONLY UPDATE WITH NEW CALL LIST
	for f := 0; f < elevio.NFloors; f++ {
		hraInput.HallRequests[f][0] = hraInput.HallRequests[f][0] || localElevator.Requests[f][elevio.BHallUp]
		hraInput.HallRequests[f][1] = hraInput.HallRequests[f][1] || localElevator.Requests[f][elevio.BHallDown]
	}

	// Save the updated HRAInput back to JSON
	return saveHRAInput(hraInput, filename)
}

func convertLocalElevatorState(localElevator elev.Elevator) (string, string, []bool) {
	// Convert behavior
	var behavior string
	switch localElevator.CurrentBehaviour {
	case elev.EBIdle:
		behavior = "idle"
	case elev.EBMoving:
		behavior = "moving"
	case elev.EBDoorOpen:
		behavior = "door open"
	}
	// Convert direction
	var direction string
	switch localElevator.Dirn {
	case elevio.DirUp:
		direction = "up"
	case elevio.DirDown:
		direction = "down"
	default:
		direction = "stop"
	}

	// Convert cab requests
	cabRequests := make([]bool, elevio.NFloors)
	for f := 0; f < elevio.NFloors; f++ {
		cabRequests[f] = localElevator.Requests[f][elevio.BCab]
	}

	return behavior, direction, cabRequests
}

func LoadHRAInput(filename string) (HRAInput, error) {
	osFile, err := filehandeling.LockFile(filename) // Lock the file for reading
	if err != nil {
		return HRAInput{}, err
	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	var hraInput HRAInput
	data, err := os.ReadFile(filename)
	if err != nil {
		return HRAInput{}, fmt.Errorf("failed to read file: %v", err)
	}
	err = json.Unmarshal(data, &hraInput)
	if err != nil {
		return HRAInput{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return hraInput, nil
}

func saveHRAInput(hraInput HRAInput, fileName string) error {
	data, err := json.MarshalIndent(hraInput, "", "  ")
	if err != nil {
		return err
	}
	osFile, err := filehandeling.LockFile(fileName)
	if err != nil {
		return err

	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func initializeHRAInput(filename string) error {
	// Create a default HRAInput. Modify this according to your requirements.
	defaultHRAInput := HRAInput{
		HallRequests: make([][2]bool, elevio.NFloors),
		States:       make(map[string]HRAElevState),
	}
	// Optionally, add default states or other initial configurations here.

	return saveHRAInput(defaultHRAInput, filename)
}

func CallCompleteToJSON(elevatorName string, filename string, localElevator elev.Elevator) error {
	hraInput, err := LoadHRAInput(filename)
	if err != nil{
		print(err)
	}

	hraInput.HallRequests[localElevator.CurrentFloor][0] = localElevator.Requests[localElevator.CurrentFloor][elevio.BHallUp]
	hraInput.HallRequests[localElevator.CurrentFloor][1] = localElevator.Requests[localElevator.CurrentFloor][elevio.BHallDown]
    hraInput.States[elevatorName].CabRequests[localElevator.CurrentFloor] = localElevator.Requests[localElevator.CurrentFloor][elevio.BCab] //false 

	return saveHRAInput(hraInput, filename)
}

func UpdateLocalElevatorToJSON(elevatorName string, filename string, localElevator elev.Elevator ) error {
    hraInput, err := LoadHRAInput(filename)
    // Convert the local elevator's state
	if err != nil{
		print(err)
	}

	behavior, direction, cabRequests := convertLocalElevatorState(localElevator)

	// Update only the specified elevator's state in HRAInput
	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       localElevator.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}

    return saveHRAInput(hraInput, filename)
}
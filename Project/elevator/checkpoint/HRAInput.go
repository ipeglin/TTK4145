package checkpoint

import (
	"elevator/elev"
	"elevator/elevio"
)

const FilenameHRAInput = "elevHRAInput.json"

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

func initializeHRAInput(el elev.Elevator, elevatorName string) HRAInput {
	// Create a default HRAInput. Modify this according to your requirements.
	hraInput := HRAInput{
		HallRequests: make([][2]bool, elevio.NFloors),
		States:       make(map[string]HRAElevState),
	}
	for f := 0; f < elevio.NFloors; f++ {
		hraInput.HallRequests[f][0] = el.Requests[f][elevio.BHallUp]
		hraInput.HallRequests[f][1] = el.Requests[f][elevio.BHallDown]
	}

	behavior, direction, cabRequests := convertLocalElevatorState(el)

	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       el.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

func updateHRAInput(hraInput HRAInput, el elev.Elevator, elevatorName string) HRAInput {
	for f := 0; f < elevio.NFloors; f++ {
		hraInput.HallRequests[f][0] = hraInput.HallRequests[f][0] || el.Requests[f][elevio.BHallUp]
		hraInput.HallRequests[f][1] = hraInput.HallRequests[f][1] || el.Requests[f][elevio.BHallDown]
	}

	behavior, direction, cabRequests := convertLocalElevatorState(el)

	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       el.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
}

func updateHRAInputWhenHallOrderIsComplete(el elev.Elevator, elevatorName string, orderCompleteFloor int) HRAInput {
	hraInput := HRAInput{
		HallRequests: make([][2]bool, elevio.NFloors),
		States:       make(map[string]HRAElevState),
	}

	hraInput.HallRequests[orderCompleteFloor][0] = el.Requests[orderCompleteFloor][elevio.BHallUp]
	hraInput.HallRequests[orderCompleteFloor][1] = el.Requests[orderCompleteFloor][elevio.BHallDown]

	behavior, direction, cabRequests := convertLocalElevatorState(el)

	hraInput.States[elevatorName] = HRAElevState{
		Behavior:    behavior,
		Floor:       el.CurrentFloor,
		Direction:   direction,
		CabRequests: cabRequests,
	}
	return hraInput
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
		behavior = "doorOpen"
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

func updateHRAInputWhenNewOrderOccurs(hraInput HRAInput, elevatorName string, btnFloor int, btn elevio.Button, localElevator *elev.Elevator) HRAInput {
	switch btn {
	case elevio.BHallUp:
		hraInput.HallRequests[btnFloor][0] = true
	case elevio.BHallDown:
		hraInput.HallRequests[btnFloor][1] = true
	case elevio.BCab:
		hraInput.States[elevatorName].CabRequests[btnFloor] = true
		localElevator.Requests[btnFloor][btn] = true
	}
	return hraInput
}

/*
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


/*
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
*/
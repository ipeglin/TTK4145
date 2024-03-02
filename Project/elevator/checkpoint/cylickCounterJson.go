package checkpoint

import (
	"elevator/elevio"
	"elevator/filehandeling"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)
type CyclicCounterState struct {
	Behavior    int   `json:"behaviour"`
	Floor       int   `json:"floor"`
	Direction   int   `json:"direction"`
	CabRequests []int `json:"cabRequests"`
}

type CyclicCounterInput struct {
	HallRequests [][2]int                      `json:"hallRequests"`
	States       map[string]CyclicCounterState `json:"states"`
}


const FilenameCylickCounter = "cylickCounter.JSON"


func SaveCyclicCounterInput(cyclicCounter CyclicCounterInput, filename string) error {
	data, err := json.MarshalIndent(cyclicCounter, "", "  ")
	if err != nil {
		return err
	}
	osFile, err := filehandeling.LockFile(filename)
	if err != nil {
		return err

	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}


func LoadCyclicCounterInput(filename string) (CyclicCounterInput, error)  {
	osFile, err := filehandeling.LockFile(filename) // Lock the file for reading
	if err != nil {
		return CyclicCounterInput{}, err
	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	var cyclicCounter CyclicCounterInput
	data, err := os.ReadFile(filename)
	if err != nil {
		return CyclicCounterInput{}, fmt.Errorf("failed to read file: %v", err)
	}
	err = json.Unmarshal(data, &cyclicCounter)
	if err != nil {
		return CyclicCounterInput{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return cyclicCounter, nil
}

func InitCyclicCounter(filename string) error { // La til returtypen error
    cyclicCounter, err := LoadCyclicCounterInput(filename) // Anta at denne funksjonen er definert et sted
    if err != nil {
        // If loading fails, try to initialize the CyclicCounter file.
        initErr := createCyclicJSON(filename) // Anta at denne funksjonen er definert et sted
        if initErr != nil {
            // If initialization also fails, return this new error.
            return fmt.Errorf("failed to initialize cyclicCounter JSON: %v", initErr)
        }
        // After initialization, attempt to load the newly created file again.
        cyclicCounter, err = LoadCyclicCounterInput(filename)
        if err != nil {
            // If loading still fails, return this error.
            return fmt.Errorf("failed to load cyclicCounter from JSON after initialization: %v", err)
        }
    }

	// Initialiser HallRequests med nuller
	for i := range cyclicCounter.HallRequests {
		cyclicCounter.HallRequests[i] = [2]int{0, 0} // Null for både opp- og ned-knapper
	}

	// Anta at du vil initialisere 'States' med en eller flere starttilstander.
	// For eksempel, for en enkelt heis med navn "one":
	cyclicCounter.States["one"] = CyclicCounterState{
		Behavior:    0,
		Floor:       0,
		Direction:   0,
		CabRequests: make([]int, elevio.NFloors),
	}

	// For å initialisere CabRequests med nuller
	for i := range cyclicCounter.States["one"].CabRequests {
		cyclicCounter.States["one"].CabRequests[i] = 0
	}

    err = SaveCyclicCounterInput(cyclicCounter, filename) // Pass på at denne funksjonen er definert et sted
    if err != nil {
        logrus.Error("Feil ved lagring av CyclicCounterInput:", err)
        return err // Endret til å returnere err direkte
    }
    return nil // Sørger for å returnere nil hvis det ikke er noen feil
}


func createCyclicJSON(filename string) error {
	// Create a default HRAInput. Modify this according to your requirements.
	defaultCyclikCounter := CyclicCounterInput{
		HallRequests: make([][2]int, elevio.NFloors),
		States:       make(map[string]CyclicCounterState),
	}
	// Optionally, add default states or other initial configurations here.

	return SaveCyclicCounterInput(defaultCyclikCounter, filename)
}
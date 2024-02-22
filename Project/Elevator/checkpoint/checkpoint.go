package checkpoint

import (
	"encoding/json"
	"fmt"
	"heislab/Elevator/elev"
	"heislab/Elevator/filehandeling"
	"os"
	"time"
)

type ElevCheckpoint struct {
	State     elev.Elevator
	Timestamp time.Time
}

const filenameCheckpoint = "elevCheckpoint.JSON" // Må finne ut av hvordan filepath burde være

func SaveElevCheckpoint(e elev.Elevator) error {

	checkpoint := ElevCheckpoint{
		State:     e,
		Timestamp: time.Now(),
	}

	// Marshal the checkpoint data to JSON
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint data: %v", err)
	}

	osFile, err := filehandeling.LockFile(filenameCheckpoint)
	if err != nil {
		return err

	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	err = os.WriteFile(filenameCheckpoint, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadElevCheckpoint() (elev.Elevator, time.Time, error) {
	osFile, err := filehandeling.LockFile(filenameCheckpoint) // Lock the file for reading
	if err != nil {
		return elev.Elevator{}, time.Time{}, err
	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	data, err := os.ReadFile(filenameCheckpoint)
	if err != nil {
		return elev.Elevator{}, time.Time{}, fmt.Errorf("failed to read checkpoint file: %v", err)
	}

	var checkpoint ElevCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return elev.Elevator{}, time.Time{}, fmt.Errorf("failed to unmarshal checkpoint data: %v", err)
	}

	return checkpoint.State, checkpoint.Timestamp, nil
}

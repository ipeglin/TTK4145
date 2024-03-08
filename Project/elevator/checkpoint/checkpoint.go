package checkpoint

import (
	"elevator/elev"
	"elevator/filehandeling"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const FilenameCheckpoint = "elevCheckpoint.JSON" // Filepath is kinda a stupid way to do this

type ElevCheckpoint struct {
	State     elev.Elevator
	Timestamp time.Time
}

func toJSON(checkpoint ElevCheckpoint) ([]byte, error) {
	return json.MarshalIndent(checkpoint, "", "  ")
}

func fromJSON(data []byte) ElevCheckpoint {
	var checkpoint ElevCheckpoint
	json.Unmarshal(data, &checkpoint)
	return checkpoint
}

func saveCheckpoint(data []byte, fileName string) error {

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

func loadCheckpoint(fileName string) ([]byte, error) {
	osFile, err := filehandeling.LockFile(fileName) // Lock the file for reading
	if err != nil {
		return nil, err
	}
	defer filehandeling.UnlockFile(osFile) // Ensure file is unlocked after reading

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %v", err)
	}
	return data, nil
}

func SaveElevCheckpoint(e elev.Elevator, fileName string) error {
	checkpoint :=
		ElevCheckpoint{
			e,
			time.Now(),
		}
	jsonCP, err := toJSON(checkpoint)
	if err != nil {
		return err
	}
	saveCheckpoint(jsonCP, fileName)
	return nil
}

func LoadElevCheckpoint(fileName string) (elev.Elevator, time.Time, error) {
	jsonCp, err := loadCheckpoint(fileName)
	if err != nil {
		return elev.Elevator{}, time.Time{}, err
	}
	checkpoint := fromJSON(jsonCp)
	return checkpoint.State, checkpoint.Timestamp, nil
}

package checkpoint

import (
	"elevator/elev"
	"elevator/filehandler"
	"encoding/json"
	"time"
)

const FilenameCheckpoint = "elevCheckpoint.json" // Filepath is kinda a stupid way to do this

type ElevCheckpoint struct {
	State     elev.Elevator
	Timestamp time.Time
}

func SaveElevCheckpoint(e elev.Elevator, fileName string) error {
	checkpoint :=
		ElevCheckpoint{
			e,
			time.Now(),
		}
	jsonCP, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}
	filehandler.WriteToFile(jsonCP, fileName)
	return nil
}

func LoadElevCheckpoint(fileName string) (elev.Elevator, time.Time, error) {
	jsonCp, err := filehandler.ReadFromFile(fileName)
	if err != nil {
		return elev.Elevator{}, time.Time{}, err
	}

	var checkpoint ElevCheckpoint
	json.Unmarshal(jsonCp, &checkpoint)
	return checkpoint.State, checkpoint.Timestamp, nil
}

package checkpoint

import (
	"elevator/elev"
	"encoding/json"
	"filehandler"
	"time"
)

const CheckpointFilename = "checkpoint.json" // Filepath is kinda a stupid way to do this

type ElevCheckpoint struct {
	State     elev.Elevator
	Timestamp time.Time
}

func SetCheckpoint(e elev.Elevator, filename string) error {
	checkpoint :=
		ElevCheckpoint{
			e,
			time.Now(),
		}

	jsonCP, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	filehandler.WriteToFile(jsonCP, filename)
	return nil
}

func LoadCheckpoint(filename string) (elev.Elevator, time.Time, error) {
	jsonCp, err := filehandler.ReadFromFile(filename)
	if err != nil {
		return elev.Elevator{}, time.Time{}, err
	}

	var checkpoint ElevCheckpoint
	json.Unmarshal(jsonCp, &checkpoint)
	return checkpoint.State, checkpoint.Timestamp, nil
}

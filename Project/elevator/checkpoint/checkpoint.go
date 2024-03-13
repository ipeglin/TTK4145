package checkpoint

import (
	"elevator/elev"
	"encoding/json"
	"filehandler"
	"time"
)

const checkpointFilename = "checkpoint.json"

type ElevCheckpoint struct {
	State     elev.Elevator
	Timestamp time.Time
}

func SetCheckpoint(e elev.Elevator) error {
	checkpoint :=
		ElevCheckpoint{
			e,
			time.Now(),
		}

	jsonCP, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	filehandler.WriteToFile(jsonCP, checkpointFilename)
	return nil
}

func LoadCheckpoint() (elev.Elevator, time.Time, error) {
	jsonCp, err := filehandler.ReadFromFile(checkpointFilename)
	if err != nil {
		return elev.Elevator{}, time.Time{}, err
	}

	var checkpoint ElevCheckpoint
	json.Unmarshal(jsonCp, &checkpoint)
	return checkpoint.State, checkpoint.Timestamp, nil
}

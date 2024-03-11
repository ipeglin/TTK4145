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

// TODO! Just a wrapper. Can be omitted
func toJSON(checkpoint ElevCheckpoint) ([]byte, error) {
	return json.MarshalIndent(checkpoint, "", "  ")
}

// TODO! Just a wrapper. Can be omitted
func fromJSON(data []byte) ElevCheckpoint {
	var checkpoint ElevCheckpoint
	json.Unmarshal(data, &checkpoint)
	return checkpoint
}

func SetElevatorCheckpoint(e elev.Elevator, fileName string) error {
	checkpoint :=
		ElevCheckpoint{
			e,
			time.Now(),
		}
	jsonCP, err := toJSON(checkpoint)
	if err != nil {
		return err
	}
	filehandler.WriteToFile(jsonCP, fileName)
	return nil
}

func LoadElevatorCheckpoint(fileName string) (elev.Elevator, time.Time, error) {
	jsonCp, err := filehandler.ReadFromFile(fileName)
	if err != nil {
		return elev.Elevator{}, time.Time{}, err
	}
	checkpoint := fromJSON(jsonCp)
	return checkpoint.State, checkpoint.Timestamp, nil
}

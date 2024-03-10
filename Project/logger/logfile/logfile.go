package logfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/log"
)

func getCurrentTimeStamp() string {
	now := time.Now()
	timestamp := fmt.Sprintf("%d-%d-%d_%d:%d:%d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Hour(),
		now.Second())

	return timestamp
}

func CreateLogFile() string {
	rootPath, err := filepath.Abs("../") // procject root
	if err != nil {
		log.Fatal("Failed to find project root", err)
	}

	timestamp := getCurrentTimeStamp()
	filename := fmt.Sprintf("%s/log/runtime_%s.log", rootPath, timestamp)

	os.MkdirAll(filepath.Dir(filename), 0755)
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	log.Info("Created log file: ", filename)

	return filename
}

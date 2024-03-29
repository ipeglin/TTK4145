package logfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
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
	rootPath, err := filepath.Abs("../")
	if err != nil {
		log.Fatal("Failed to find project root", err)
	}

	timestamp := getCurrentTimeStamp()
	filename := fmt.Sprintf("runtime_%s.log", timestamp)
	path := fmt.Sprintf("%s/tmp/%s", rootPath, filename)

	os.MkdirAll(filepath.Dir(path), 0755)
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.Info("Created log file: ", filename)

	return path
}

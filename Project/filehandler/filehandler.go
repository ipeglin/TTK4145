package filehandler

import (
	"fmt"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

func WriteToFile(data []byte, filename string) error {
	osFile, err := lockFile(filename)
	if err != nil {
		return err
	}
	defer unlockFile(osFile)

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	logrus.Debug("Successfully wrote to file: ", filename)
	return nil
}

func ReadFromFile(filename string) ([]byte, error) {
	osFile, err := lockFile(filename)
	if err != nil {
		return nil, err
	}
	defer unlockFile(osFile)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %v", err)
	}
	logrus.Debug("Successfully read from file: ", filename)
	return data, nil
}

func lockFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return nil, err
	}
	logrus.Debug("Locked file: ", filename)
	return file, nil
}

func unlockFile(file *os.File) error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	logrus.Debug("Unlocked file: ", file.Name())
	return closeErr
}

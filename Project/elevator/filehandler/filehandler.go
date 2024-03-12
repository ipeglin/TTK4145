package filehandler

import (
	"fmt"
	"os"
	"syscall"
)

func WriteToFile(data []byte, fileName string) error {
	osFile, err := lockFile(fileName)
	if err != nil {
		return err

	}
	defer unlockFile(osFile) // Ensure file is unlocked after reading

	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ReadFromFile(fileName string) ([]byte, error) {
	osFile, err := lockFile(fileName) // Lock the file for reading
	if err != nil {
		return nil, err
	}
	defer unlockFile(osFile) // Ensure file is unlocked after reading

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %v", err)
	}
	return data, nil
}

func lockFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func unlockFile(file *os.File) error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

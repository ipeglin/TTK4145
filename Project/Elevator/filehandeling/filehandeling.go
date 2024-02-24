package filehandeling

import (
	"os"
	"syscall"
)

func LockFile(filePath string) (*os.File, error) {
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

func UnlockFile(file *os.File) error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

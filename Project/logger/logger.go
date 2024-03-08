package logger

import (
  "os"
  "io"
  "fmt"
  "time"
  "path/filepath"
  "github.com/sirupsen/logrus"
)

func createLogFile() string {
  rootPath, err := filepath.Abs("../") // procject root
  if err != nil {
      logrus.Fatal("Failed to find project root", err)
  }

  // generate timestamp
  now := time.Now()
  timestamp := fmt.Sprintf("runtime_%d-%d-%d_%d:%d:%d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Hour(),
		now.Second())

  filename := fmt.Sprintf("%s/log/%s.log", rootPath, timestamp)
  os.MkdirAll(filepath.Dir(filename), 0755)
  file, err := os.Create(filename)
      if err != nil {
          logrus.Fatal(err)
      }
  file.Close()
  logrus.Info("Created log file: ", filename)

  return filename
}

func Setup() {
  filePath := createLogFile()

  // pass log file to logrus
  file, err := os.OpenFile(filePath, os.O_WRONLY | os.O_CREATE, 0755)
  if err != nil {
      logrus.Fatal("Failed to create log file. ", err)
  }
  mw := io.MultiWriter(os.Stdout, file)
  logrus.SetOutput(mw)
  logrus.SetReportCaller(true)
  logrus.SetLevel(logrus.DebugLevel)
}

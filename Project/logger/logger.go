package logger

import (
	"fmt"
	"io"
	"logger/logfile"
	"os"

	log "github.com/sirupsen/logrus"
)

/*
Hook code for writing log messages to separate targets based on log level

author: @bunyk
*/
type WriterHook struct {
	Writer    io.Writer
	LogLevels []log.Level
}

func (hook *WriterHook) Fire(entry *log.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

func (hook *WriterHook) Levels() []log.Level {
	return hook.LogLevels
}

func Setup() {
	file := logfile.CreateLogFile()
	fmt.Println("HERE:", file)

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("Failed to create log file. ", err)
	}

	log.SetOutput(io.Discard) // Send all logs to nowhere by default
	log.AddHook(&WriterHook{  // Send logs with level higher than warning to log file
		Writer: f,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
			log.DebugLevel,
		},
	})

	log.AddHook(&WriterHook{ // Send info and trace logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.TraceLevel,
		},
	})
}

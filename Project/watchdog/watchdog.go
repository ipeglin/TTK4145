package watchdog

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
)

/* Pseudo Code
1. If pid is not 0, then the process is a backup process.
	A. Loop to check if main process PID is still running
	B. If not, break the loop

2. Is main process, or should claim itself as new primary process
	A. Start backup process watching itself
	B. Wait for termination signal
	C. Terminate backup process and then itself, if signal received
*/
func Init(pid int, done chan<- bool) {

	// commence watching
	if pid != 0 {
		logrus.Info("Watchdog overlooking process ", pid)

		for {
			p, err := ps.FindProcess(pid)
			if err != nil {
				logrus.Error("Error:", err)
			}
			if p == nil {
				logrus.Warn("Primary process has terminated")
				break
			}

			time.Sleep(500 * time.Millisecond)
		}
	}

	// assume primary process
	pid = os.Getpid()
	logrus.Debug("No process to watch. Selfassigning primary process:", pid)

	// fetch init entrypoints relative to pwd (must be TTK4145/Project)
	entrypoint, err := filepath.Abs("./init/init.go")
	logrus.Debug("Initialisation entrypoint:", entrypoint, err)
	if err != nil {
		logrus.Fatal("Error getting entrypoint:", err)
	}

	// non-blocking process start
	runCmd := cmd.NewCmd("go", "run", entrypoint, "-watch", fmt.Sprintf("%d", pid))
	runCmd.Start() // non-blocking

	// await fetching PID of child process
	var childProcessPid string
	receivedPid := make(chan bool, 1)
	go func() {
		<-time.After(500 * time.Millisecond)
		status := runCmd.Status()
		n := len(status.Stdout)
		childProcessPid = status.Stdout[n-1]
		logrus.Debug("Catching watchdog PID: ", childProcessPid)
		receivedPid <- true
	}()

	// wait for PID to be received
	<-receivedPid

	logrus.Info("Watchdog process ", childProcessPid, " is watching over process ", pid)
	done <- true // signal continuation to init module

	// gracefully handle termination
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// await termination
	termination := make(chan bool, 1)
	go func() {
		sig := <-sigs
		logrus.Debug("Received Signal: ", sig)
		termination <- true
	}()

	<-termination

	// terminate processes
	runCmd.Stop()

	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", os.Getpid())).Run()
	} else {
		os.Exit(0)
	}
}
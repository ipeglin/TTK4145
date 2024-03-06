package watchdog

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
)

func Init(pid int, done chan<- bool) {

	/* Pseudo Code
	1. If pid is not 0, then the process is a backup process.
		A. Loop to check if main process PID is still running
		B. If not, break the loop

	2. Is main process, or should claim itself as new primary process
		A. Start backup process watching itself
		B. Wait for termination signal
		C. Terminate backup process and then itself, if signal received
	*/

	/* Code begins here */

	// 1. If pid is not 0, then the process is a backup process.
	if pid != 0 {
		logrus.Warn("Watchdog overlooking process ", pid)

		// A. Loop to check if main process PID is still running
		for {
			p, err := ps.FindProcess(pid)
			if err != nil {
				logrus.Warn("Error:", err)
			}
			if p == nil {
				// B. If not, break the loop
				logrus.Warn("Primary process has terminated")
				break
			}

			done <- true
			time.Sleep(1 * time.Second)
		}
	}

	// 2. Is main process, or should claim itself as new primary process
	pid = os.Getpid()
	logrus.Debug("No process to watch. Selfassigning primary process:", pid)

	// A. Start backup process watching itself
	entrypoint, err := filepath.Abs("./init/init.go")
	logrus.Debug("Initialisation entrypoint:", entrypoint, err)
	if err != nil {
		logrus.Fatal("Error getting entrypoint:", err)
	}

	cmd := exec.Command("go", "run", entrypoint, "-watch", fmt.Sprintf("%d", pid))
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// Capture stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
			logrus.Info("Error creating StdoutPipe:", err)
			return
	}

	logrus.Info("Starting watchdog process to overlook process: ", pid)
	err = cmd.Start()
	if err != nil {
		logrus.Fatal("Failed to start backup process:", err)
	}

	err = cmd.Process.Release()
	if err != nil {
			logrus.Fatal("cmd.Process.Release failed: ", err)
	}


	// Read the output line by line
	var childProcess int
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
			line := scanner.Text()
			// Assuming the line contains the PID, you might adjust this based on your program's output
			childPidStr := strings.TrimSpace(line)
			childProcess, err = strconv.Atoi(childPidStr)
			if err != nil {
					logrus.Info("Error converting PID:", err)
					return
			}
			logrus.Info("PID of the spawned process:", childProcess)
			// Now you have the PID of the spawned process, you can do further operations with it.
	}

	// Check for any errors
	if err := scanner.Err(); err != nil {
			logrus.Info("Error reading stdout:", err)
			return
	}

	logrus.Warn("Watchdog cmd exec PID:", cmd.Process.Pid)
	done <- true

	// gracefully handle termination
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	termination := make(chan bool, 1)
	go func() {
		sig := <-sigs
		fmt.Println("HERE!!! ", sig)
		termination <- true
	}()

	// B. Wait for termination signal
	<-termination

	// C. Terminate backup process and then itself, if signal received
	// Find the process by its PID
	p, err := ps.FindProcess(childProcess)
	if err != nil {
			fmt.Println("Error finding process:", err)
			return
	}

	// Kill the process
	fmt.Println("Watchdog Process: ", p)
	// err = syscall.Kill(process.Pid, syscall.SIGKILL)
	// if err != nil {
	// 		fmt.Println("Error killing process:", err)
	// 		return
	// }

	logrus.Warn("Process with PID", childProcess, "has been killed.")
}
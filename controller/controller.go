package controller

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	tracker "github.com/crimist/trakx/cmd"
	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"github.com/pkg/errors"
)

const (
	logFilePermissions = 0644
)

type Controller struct {
	processIDFile *ProcessIDFile
	config        *config.Configuration
}

func NewController(conf *config.Configuration) *Controller {
	return &Controller{
		processIDFile: NewProcessIDFile(conf.PIDPath()),
		config:        conf,
	}
}

// Execute executes trakx in the current process
func (controller *Controller) Execute() {
	tracker.Run(controller.config)
}

// Start starts trakx as a service
func (controller *Controller) Start() error {
	pidFileExists, processAlive, heartbeat := controller.Status()
	if pidFileExists || processAlive || heartbeat {
		return errors.New("trakx is already running")
	}

	logFile, err := os.OpenFile(controller.config.LogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFilePermissions)
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}
	defer logFile.Close()

	cmd := exec.Command(os.Args[0], "execute")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start trakx process")
	}

	if err := controller.processIDFile.Write(cmd.Process.Pid); err != nil {
		return errors.Wrap(err, "failed to write process id to file")
	}

	fmt.Println("started trakx!")
	return nil
}

// Stop gracefully stops the process by sending a stop signal
func (controller *Controller) Stop() error {
	process, err := controller.processIDFile.Process()
	if err != nil {
		return errors.Wrap(err, "failed to get process from process id file")
	}
	if err := process.Signal(tracker.SigStop); err != nil {
		return errors.Wrap(err, "failed to send stop signal to process")
	}

	processid, err := controller.processIDFile.Read()
	if err != nil {
		return errors.Wrap(err, "failed to read process id file")
	}

	fmt.Print("waiting for process to exit")
	i := 0
	for ; err == nil || i == 100; err = syscall.Kill(processid, syscall.Signal(0)) {
		time.Sleep(100 * time.Millisecond)
		fmt.Print(".")
		i++
	}
	if i == 100 {
		return errors.New(" trakx failed to stop within 10s")
	}
	if err.Error() != "no such process" {
		return errors.Wrap(err, "failed to kill trakx process id")
	}

	fmt.Println(" stopped trakx!")
	return errors.Wrap(controller.processIDFile.Clear(), "failed to clear trakx process id file")
}

// Clear clears the trakx process id file
func (controller *Controller) Clear() error {
	return errors.Wrap(controller.processIDFile.Clear(), "failed to clear trakx process id file")
}

// Status returns the status of trakx by checking the following:
// process id file exists, process id file has pid, proces is alive, heartbeat to trakx
func (controller *Controller) Status() (pidFileExists bool, processAlive bool, heartbeat bool) {
	// check process id file exists
	processid, _ := controller.processIDFile.Read()
	if processid != ProcessIDFailed {
		pidFileExists = true

		// check process is alive
		if err := syscall.Kill(processid, syscall.Signal(0)); err == nil {
			processAlive = true
		}
	}

	// heartbeat checks
	if controller.config.UDP.Enabled {
		conn, err := net.Dial("udp", fmt.Sprintf("localhost:%d", controller.config.UDP.Port))
		if err == nil {
			conn.Write(udpprotocol.HeartbeatRequest)
			data := make([]byte, 1)
			size, err := conn.Read(data)

			if err == nil {
				if size == 1 && bytes.Equal(data, udpprotocol.HeartbeatOk) {
					heartbeat = true
				}
			}
		}
	}

	if controller.config.HTTP.Mode == config.TrackerModeEnabled {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/heartbeat", controller.config.HTTP.Port))
		if err == nil && resp.StatusCode == 200 {
			heartbeat = true
		}
	}

	return
}

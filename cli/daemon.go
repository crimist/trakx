package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/daemon"
	"github.com/crimist/trakx/internal/pidfile"
	"github.com/crimist/trakx/tracker/udp/udpprotocol"
	"github.com/pkg/errors"
)

const (
	logFilePermissions = 0644
)

type daemonController struct {
	processIDFile *pidfile.File
	config        *config.Configuration
}

func newDaemonController(conf *config.Configuration) *daemonController {
	return &daemonController{
		processIDFile: pidfile.New(conf.PIDPath()),
		config:        conf,
	}
}

func (controller *daemonController) Execute() {
	daemon.Run(controller.config)
}

func (controller *daemonController) Start(opts GlobalOptions, importPath string) error {
	pidFileExists, processAlive, heartbeat := controller.Status()
	if pidFileExists || processAlive || heartbeat {
		return errors.New("trakx is already running")
	}

	logFile, err := os.OpenFile(controller.config.LogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFilePermissions)
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}
	defer logFile.Close()

	args := []string{"run"}
	if opts.ConfigPath != "" {
		args = append(args, "--config", opts.ConfigPath)
	}
	if importPath != "" {
		args = append(args, "--import", importPath)
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start trakx process")
	}

	if err := controller.processIDFile.Write(cmd.Process.Pid); err != nil {
		return errors.Wrap(err, "failed to write process id to file")
	}

	return nil
}

func (controller *daemonController) Stop(out io.Writer) error {
	process, err := controller.processIDFile.Process()
	if err != nil {
		return errors.Wrap(err, "failed to get process from process id file")
	}
	if err := process.Signal(daemon.SigStop); err != nil {
		return errors.Wrap(err, "failed to send stop signal to process")
	}

	processid, err := controller.processIDFile.Read()
	if err != nil {
		return errors.Wrap(err, "failed to read process id file")
	}

	fmt.Fprint(out, "waiting for process to exit")
	i := 0
	for pidfile.IsAlive(processid) && i < 100 {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(out, ".")
		i++
	}
	fmt.Fprintln(out)
	if i == 100 {
		return errors.New("trakx failed to stop within 10s")
	}

	return errors.Wrap(controller.processIDFile.Clear(), "failed to clear trakx process id file")
}

func (controller *daemonController) Clear() error {
	return errors.Wrap(controller.processIDFile.Clear(), "failed to clear trakx process id file")
}

func (controller *daemonController) Status() (pidFileExists bool, processAlive bool, heartbeat bool) {
	processid, _ := controller.processIDFile.Read()
	if processid != pidfile.ProcessIDFailed {
		pidFileExists = true

		if pidfile.IsAlive(processid) {
			processAlive = true
		}
	}

	if controller.config.UDP.Port != 0 {
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

	if controller.config.HTTP.Tracker {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/heartbeat", controller.config.HTTP.Port))
		if err == nil && resp.StatusCode == 200 {
			heartbeat = true
		}
	}

	return
}

package controller

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/crimist/trakx/tracker"
	"github.com/crimist/trakx/tracker/shared"
	"github.com/pkg/errors"
)

const (
	perms   = 0740
	pidFile = shared.TrakxRoot + "trakx.pid"
	logFile = shared.TrakxRoot + "trakx.log"
)

type controller struct {
	permissions uint
	pID         *pID
	logPath     string
}

func NewController() *controller {
	c := &controller{
		permissions: perms,
		pID:         newPID(pidFile, perms),
		logPath:     logFile,
	}

	return c
}

// Run runs trakx
func (c *controller) Run() {
	fmt.Println("Running...")
	tracker.Run()
	fmt.Println("Ran!")
}

// Start starts trakx as a service
func (c *controller) Start() error {
	fmt.Println("starting...")
	if c.Running() {
		return errors.New("Trakx is already running")
	}

	logFile, err := os.OpenFile(c.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(c.permissions))
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}
	defer logFile.Close()

	cmd := exec.Command(os.Args[0], "run")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start process")
	}

	if err := c.pID.write(cmd.Process.Pid); err != nil {
		return errors.Wrap(err, "failed to write pid to file")
	}

	fmt.Println("started!")
	return nil
}

// Stop stops the trakx service
func (c *controller) Stop() error {
	fmt.Println("stopping...")

	process, err := c.pID.Process()
	if err != nil {
		return err
	}
	if err := process.Signal(tracker.SigStop); err != nil {
		return err
	}
	if err := process.Release(); err != nil {
		return err
	}

	pid, err := c.pID.read()
	if err != nil {
		return err
	}

	// process.Wait() fails on most systems since it's not a child
	// As such we just have to keep checking if the process died

	start := time.Now()
	for ; err == nil; err = syscall.Kill(pid, syscall.Signal(0)) {
		fmt.Println("Waiting for death", time.Since(start))
		time.Sleep(50 * time.Millisecond)
	}
	if err.Error() != "no such process" {
		return err
	}

	fmt.Println("stopped!")
	return c.pID.clear()
}

// Wipe clears the trakx pid file
func (c *controller) Wipe() error {
	return c.pID.clear()
}

// Running checks if trakx is running using bind
func (c *controller) Running() bool {
	config, err := shared.LoadConf(nil)
	if err != nil {
		panic(err) // TODO: handle
	}

	if config.Tracker.UDP.Enabled {
		conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: config.Tracker.UDP.Port, Zone: ""})
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") {
				return true
			}
		} else {
			conn.Close()
		}
	}

	if config.Tracker.HTTP.Enabled {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/announce", config.Tracker.HTTP.Port))
		if err == nil && resp.StatusCode == 200 {
			return true
		}
	}

	return false
}

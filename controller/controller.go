package controller

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/crimist/trakx/tracker"
)

type Controller struct {
	rootPerms uint
	pID       *pID
	logPath   string
}

// NewController creates a controller with root at given dir
func NewController(root string, perms os.FileMode) (*Controller, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	root = strings.Replace(root, "~", home, 1)

	c := &Controller{}
	c.pID = newpID(root+"trakx.pid", perms)
	c.logPath = root + "trakx.log"

	oldMask := syscall.Umask(0)
	if err := os.MkdirAll(root, 0740); err != nil {
		panic(err)
	}
	syscall.Umask(oldMask)

	return c, nil
}

// Run runs trakx
func (c *Controller) Run() {
	fmt.Println("Running...")
	tracker.Run()
	fmt.Println("Ran!")
}

// Start starts trakx as a service
func (c *Controller) Start() error {
	fmt.Println("starting...")
	if c.IsRunning() {
		return errors.New("Trakx is already running")
	}

	logFile, err := os.OpenFile(c.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(c.rootPerms))
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(os.Args[0], "run")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Println("started!")
	return c.pID.write(cmd.Process.Pid)
}

// Stop stops the trakx service
func (c *Controller) Stop() error {
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
func (c *Controller) Wipe() error {
	return c.pID.clear()
}

// IsRunning checks if trakx is running using bind
func (c *Controller) IsRunning() (running bool) {
	if conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 1337, Zone: ""}); err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			running = true
		} else {
			panic(err)
		}
	} else {
		conn.Close()
	}
	return
}

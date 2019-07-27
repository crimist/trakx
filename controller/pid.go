package controller

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const trakxNotRunning = -1

type pID struct {
	path  string
	perms os.FileMode
}

func NewpID(path string, perms os.FileMode) *pID {
	p := &pID{}
	p.path = path
	p.perms = perms

	return p
}

// Read gets the pid. Returns -1 if there's no pid.
func (p *pID) Read() (int, error) {
	data, err := ioutil.ReadFile(p.path)
	if os.IsNotExist(err) || string(data) == "" {
		return trakxNotRunning, nil
	} else if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func (p *pID) Write(pid int) error {
	return ioutil.WriteFile(p.path, []byte(fmt.Sprintf("%d", pid)), p.perms)
}

func (p *pID) Clear() error {
	file, err := os.OpenFile(p.path, os.O_CREATE|os.O_RDWR, p.perms)
	if err != nil {
		return err
	}
	file.Truncate(0)
	file.Seek(0, 0)
	return nil
}

func (p *pID) Process() (*os.Process, error) {
	pid, err := p.Read()
	if err != nil {
		return nil, err
	} else if pid == trakxNotRunning {
		return nil, errors.New("Trakx isn't running")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	return process, nil
}

package controller

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const notRunning = -1

type pID struct {
	path  string
	perms os.FileMode
}

func newPID(path string, perms os.FileMode) *pID {
	p := &pID{
		path:  path,
		perms: perms,
	}

	return p
}

// read gets the pid. Returns -1 if there's no pid.
func (p *pID) read() (int, error) {
	data, err := ioutil.ReadFile(p.path)
	if os.IsNotExist(err) || string(data) == "" {
		return notRunning, nil
	} else if err != nil {
		return 0, errors.Wrap(err, "failed to read pid file")
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse pid file")
	}
	return pid, nil
}

func (p *pID) write(pid int) error {
	return ioutil.WriteFile(p.path, []byte(fmt.Sprintf("%d", pid)), p.perms)
}

func (p *pID) clear() error {
	file, err := os.OpenFile(p.path, os.O_CREATE|os.O_RDWR, p.perms)
	if err != nil {
		return errors.Wrap(err, "failed to truncate the pid file")
	}
	file.Truncate(0)
	file.Seek(0, 0)
	return nil
}

func (p *pID) Process() (*os.Process, error) {
	pid, err := p.read()
	if err != nil {
		return nil, err
	} else if pid == notRunning {
		return nil, errors.New("Trakx isn't running")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find process with pid")
	}
	return process, nil
}

package main

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	pidFilePermissions = 0644
	processIDFailed    = -1
)

var (
	errFileEmpty   = errors.New("process id file is empty")
	errParseFailed = errors.New("failed to parse process id file")
)

type processIDFile struct {
	path string
}

func newProcessIDFile(path string) *processIDFile {
	return &processIDFile{path: path}
}

func (pidFile *processIDFile) Read() (int, error) {
	contents, err := os.ReadFile(pidFile.path)
	if err != nil {
		return processIDFailed, err
	}
	if len(contents) == 0 {
		return processIDFailed, errFileEmpty
	}

	processID, err := strconv.Atoi(string(contents))
	if err != nil {
		return processIDFailed, errParseFailed
	}

	return processID, nil
}

func (pidFile *processIDFile) Write(processID int) error {
	return os.WriteFile(pidFile.path, []byte(strconv.Itoa(processID)), pidFilePermissions)
}

func (pidFile *processIDFile) Clear() error {
	return os.Truncate(pidFile.path, 0)
}

func (pidFile *processIDFile) Process() (*os.Process, error) {
	processid, err := pidFile.Read()
	if err != nil {
		return nil, err
	}

	process, err := os.FindProcess(processid)
	if err != nil {
		return nil, err
	}
	return process, nil
}

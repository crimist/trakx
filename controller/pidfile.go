package controller

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	pidFilePermissions = 0644
	ProcessIDFailed    = -1
)

var (
	ErrFileEmpty   = errors.New("process id file is empty")
	ErrParseFailed = errors.New("failed to parse process id file")
)

type ProcessIDFile struct {
	path string
}

func NewProcessIDFile(path string) *ProcessIDFile {
	return &ProcessIDFile{
		path: path,
	}
}

// Read returns the process ID integer stored in the file
// returns ProcessIDFailed and the error associated if the read fails
func (pidFile *ProcessIDFile) Read() (int, error) {
	contents, err := os.ReadFile(pidFile.path)
	if err != nil {
		return ProcessIDFailed, err
	}
	if len(contents) == 0 {
		return ProcessIDFailed, ErrFileEmpty
	}

	processID, err := strconv.Atoi(string(contents))
	if err != nil {
		return ProcessIDFailed, ErrParseFailed
	}

	return processID, nil
}

// Write writes the given process ID integer to the process ID file
func (pidFile *ProcessIDFile) Write(processID int) error {
	return os.WriteFile(pidFile.path, []byte(strconv.Itoa(processID)), pidFilePermissions)
}

// Clear clears the process id by truncating the process id file
func (pidFile *ProcessIDFile) Clear() error {
	return os.Truncate(pidFile.path, 0)
}

// Process returns an os.Process for the process id in the process id file
func (pidFile *ProcessIDFile) Process() (*os.Process, error) {
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

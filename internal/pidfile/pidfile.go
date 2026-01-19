package pidfile

import (
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

const (
	PIDFilePermissions = 0644
	ProcessIDFailed    = -1
)

var (
	ErrFileEmpty   = errors.New("process id file is empty")
	ErrParseFailed = errors.New("failed to parse process id file")
)

// File represents a PID file for reading/writing process IDs.
// It is safe for concurrent use.
type File struct {
	path string
	mu   sync.RWMutex
}

// New creates a new PID file handler
func New(path string) *File {
	return &File{path: path}
}

// Read reads the process ID from the file
func (f *File) Read() (int, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	contents, err := os.ReadFile(f.path)
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

// Write writes a process ID to the file
func (f *File) Write(processID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return os.WriteFile(f.path, []byte(strconv.Itoa(processID)), PIDFilePermissions)
}

// Clear truncates the PID file
func (f *File) Clear() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return os.Truncate(f.path, 0)
}

// Process returns an os.Process for the PID in the file
func (f *File) Process() (*os.Process, error) {
	processID, err := f.Read()
	if err != nil {
		return nil, err
	}

	process, err := os.FindProcess(processID)
	if err != nil {
		return nil, err
	}
	return process, nil
}

// IsAlive checks if a process with the given PID is running.
// Returns true if the process exists, false otherwise.
// Handles EPERM (permission denied) as "alive" since the process exists.
func IsAlive(processID int) bool {
	if processID <= 0 {
		return false
	}
	err := syscall.Kill(processID, syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

// IsAlive reads the PID from the file and checks if that process is alive
func (f *File) IsAlive() (bool, error) {
	processID, err := f.Read()
	if err != nil {
		return false, err
	}
	return IsAlive(processID), nil
}

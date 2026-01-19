package backup

import (
	"io"
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Manager orchestrates backup operations across different transport mechanisms.
type Manager struct {
	backupFilePath string
	pidFilePath    string
	cacheDir       string

	// For daemon-side operations
	provider SnapshotProvider
}

// Config holds configuration for the backup manager.
type Config struct {
	BackupFilePath string
	PIDFilePath    string
	CacheDir       string
	Provider       SnapshotProvider // For daemon-side snapshot creation
}

// NewManager creates a new backup manager.
func NewManager(cfg Config) *Manager {
	return &Manager{
		backupFilePath: cfg.BackupFilePath,
		pidFilePath:    cfg.PIDFilePath,
		cacheDir:       cfg.CacheDir,
		provider:       cfg.Provider,
	}
}

// Export reads a backup from the best available source and writes it to the destination.
// Strategy: Try daemon socket first (if daemon running), fall back to file.
func (m *Manager) Export(dest Destination) error {
	source, err := m.selectExportSource()
	if err != nil {
		return err
	}

	zap.L().Info("Exporting backup",
		zap.String("source", source.Description()),
		zap.String("destination", dest.Description()))

	return m.copy(source, dest)
}

// Import reads a backup from source and writes it to the configured backup file.
func (m *Manager) Import(source Source) error {
	if m.backupFilePath == "" {
		return errors.New("backup file path not configured")
	}

	dest := NewFileDestination(m.backupFilePath)

	zap.L().Info("Importing backup",
		zap.String("source", source.Description()),
		zap.String("destination", dest.Description()))

	return m.copy(source, dest)
}

// Persist writes a snapshot from the daemon's current state.
// Strategy: Try socket first (for CLI consumption), fall back to file.
func (m *Manager) Persist() error {
	if m.provider == nil {
		return errors.New("no snapshot provider configured")
	}

	// Try socket first (if CLI is listening)
	socketDest := NewSocketDestination(os.Getpid(), m.cacheDir)
	if err := m.writeSnapshot(socketDest); err == nil {
		zap.L().Debug("Persisted backup via socket")
		return nil
	} else {
		zap.L().Debug("Socket unavailable, falling back to file", zap.Error(err))
	}

	// Fall back to file
	if m.backupFilePath == "" {
		return errors.New("backup file path not configured")
	}

	fileDest := NewFileDestination(m.backupFilePath)
	zap.L().Debug("Persisting backup to file", zap.String("path", m.backupFilePath))
	return m.writeSnapshot(fileDest)
}

// RestoreCombined is a convenience wrapper around RestoreCombined
// that works with Source interface.
func (m *Manager) RestoreCombined(source Source) (connData []byte, dbReader io.ReadCloser, err error) {
	reader, err := source.Open()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open source")
	}
	defer func() {
		if err != nil {
			reader.Close()
		}
	}()

	return RestoreCombined(reader)
}

// selectExportSource chooses the best source for export.
// Prefers live daemon snapshot over stale file.
func (m *Manager) selectExportSource() (Source, error) {
	// Try daemon socket first
	processID, err := m.readProcessID()
	if err == nil && m.isProcessAlive(processID) {
		zap.L().Debug("Using daemon socket for export", zap.Int("pid", processID))
		return NewSocketSource(processID, m.cacheDir), nil
	}

	if err != nil {
		zap.L().Debug("No daemon process found", zap.Error(err))
	}

	// Fall back to file
	if m.backupFilePath == "" {
		return nil, errors.New("backup file path not configured")
	}

	zap.L().Debug("Using file for export", zap.String("path", m.backupFilePath))
	return NewFileSource(m.backupFilePath), nil
}

// copy copies data from source to destination.
func (m *Manager) copy(source Source, dest Destination) error {
	reader, err := source.Open()
	if err != nil {
		return errors.Wrap(err, "failed to open source")
	}
	defer reader.Close()

	writer, err := dest.Create()
	if err != nil {
		return errors.Wrap(err, "failed to create destination")
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "failed to close destination")
		}
	}()

	if _, err := io.Copy(writer, reader); err != nil {
		return errors.Wrap(err, "failed to copy backup data")
	}

	return nil
}

// writeSnapshot writes a combined snapshot using the configured provider.
func (m *Manager) writeSnapshot(dest Destination) error {
	writer, err := dest.Create()
	if err != nil {
		return errors.Wrap(err, "failed to create destination")
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "failed to close destination")
		}
	}()

	db, connDB := m.provider.GetSnapshottables()
	if err := WriteCombined(writer, db, connDB); err != nil {
		return errors.Wrap(err, "failed to write combined snapshot")
	}

	return nil
}

// readProcessID reads the daemon PID from file.
func (m *Manager) readProcessID() (int, error) {
	if m.pidFilePath == "" {
		return -1, errors.New("PID file path not configured")
	}

	zap.L().Debug("Reading process ID file", zap.String("path", m.pidFilePath))
	contents, err := os.ReadFile(m.pidFilePath)
	if err != nil {
		return -1, err
	}
	if len(contents) == 0 {
		return -1, errors.New("PID file is empty")
	}

	processID, err := strconv.Atoi(string(contents))
	if err != nil {
		return -1, errors.New("failed to parse PID file")
	}

	return processID, nil
}

// isProcessAlive checks if a process is running.
func (m *Manager) isProcessAlive(processID int) bool {
	if processID <= 0 {
		return false
	}
	zap.L().Debug("Checking process liveness", zap.Int("pid", processID))
	err := syscall.Kill(processID, syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

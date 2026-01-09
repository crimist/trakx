package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage/database"
	"github.com/crimist/trakx/tracker/udp/connections"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	backupFilePermissions = 0640 // rw-r-----
	backupSocketPrefix    = "trakx-backup-"
	backupAcceptTimeout   = 5 * time.Second
	backupStreamTimeout   = 2 * time.Minute
	backupDialTimeout     = 500 * time.Millisecond
)

// ExportBackup requests a fresh snapshot from the running daemon if present,
// otherwise it streams the on-disk backup file.
func ExportBackup(conf *config.Configuration, writer io.Writer) error {
	processID, err := readProcessID(conf.PIDPath())
	if err == nil && isProcessAlive(processID) {
		zap.L().Debug("Exporting backup via daemon socket", zap.Int("pid", processID))
		return streamSnapshotFromDaemon(processID, writer)
	}
	if err != nil {
		zap.L().Debug("No daemon process found for backup export", zap.Error(err))
	}

	zap.L().Debug("Exporting backup via file", zap.String("path", conf.DB.Backup.Path))
	return exportBackupFile(conf, writer)
}

// ImportBackup validates and writes snapshot data from the reader into the configured backup file.
func ImportBackup(conf *config.Configuration, reader io.Reader) (err error) {
	backupPath := conf.DB.Backup.Path
	if backupPath == "" {
		return errors.New("backup path is empty")
	}

	zap.L().Debug("Importing backup from stream", zap.String("path", backupPath))

	tmpDir := filepath.Dir(backupPath)
	tmpFile, err := os.CreateTemp(tmpDir, "trakx-db-import-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp backup file")
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	validationDB, err := database.NewDatabase(database.Config{
		Collector: stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		return errors.Wrap(err, "failed to initialize validation database")
	}

	tee := io.TeeReader(reader, tmpFile)
	connSnapshot, dbReader, err := splitCombinedSnapshot(tee)
	if err != nil {
		return errors.Wrap(err, "failed to parse combined snapshot")
	}
	if err = validationDB.Restore(dbReader); err != nil {
		return errors.Wrap(err, "database validation failed")
	}
	if len(connSnapshot) > 0 {
		connValidation := connections.NewConnections(0, time.Minute, 0)
		if err := connValidation.Unmarshal(connSnapshot); err != nil {
			return errors.Wrap(err, "connection validation failed")
		}
	}
	if err = tmpFile.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync backup file")
	}
	if err = tmpFile.Close(); err != nil {
		return errors.Wrap(err, "failed to close backup file")
	}
	if err = os.Rename(tmpPath, backupPath); err != nil {
		return errors.Wrap(err, "failed to replace backup file")
	}
	if err = os.Chmod(backupPath, backupFilePermissions); err != nil {
		return errors.Wrap(err, "failed to set backup file permissions")
	}

	return nil
}

func persistSnapshot(db *database.Database, connDB *connections.Connections, backupPath string) error {
	if err := streamSnapshotToSocket(db, connDB); err == nil {
		zap.L().Debug("Persisted backup via socket stream")
		return nil
	}
	if backupPath == "" {
		return errors.New("backup path is empty")
	}

	zap.L().Debug("Persisting backup via file", zap.String("path", backupPath))
	return writeCombinedSnapshotFile(db, connDB, backupPath)
}

func streamSnapshotFromDaemon(processID int, writer io.Writer) error {
	socketPath := backupSocketPath(processID)
	zap.L().Debug("Preparing backup socket listener", zap.String("path", socketPath))
	if err := removeSocketPath(socketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return errors.Wrap(err, "failed to listen on backup socket")
	}
	defer listener.Close()
	defer removeSocketPath(socketPath)

	if unixListener, ok := listener.(*net.UnixListener); ok {
		if err := unixListener.SetDeadline(time.Now().Add(backupAcceptTimeout)); err != nil {
			return errors.Wrap(err, "failed to set backup accept timeout")
		}
	}

	zap.L().Debug("Signaling daemon for backup", zap.Int("pid", processID))
	if err := syscall.Kill(processID, syscall.SIGUSR1); err != nil {
		return errors.Wrap(err, "failed to signal daemon for backup")
	}

	zap.L().Debug("Waiting for daemon backup connection", zap.Duration("timeout", backupAcceptTimeout))
	conn, err := listener.Accept()
	if err != nil {
		return errors.Wrap(err, "failed to accept daemon backup connection")
	}
	defer conn.Close()

	if unixConn, ok := conn.(*net.UnixConn); ok {
		if err := unixConn.SetDeadline(time.Now().Add(backupStreamTimeout)); err != nil {
			return errors.Wrap(err, "failed to set backup stream timeout")
		}
	}

	zap.L().Debug("Streaming daemon backup to writer")
	if _, err := io.Copy(writer, conn); err != nil {
		return errors.Wrap(err, "failed to stream daemon backup")
	}

	return nil
}

func streamSnapshotToSocket(db *database.Database, connDB *connections.Connections) error {
	socketPath := backupSocketPath(os.Getpid())
	zap.L().Debug("Dialing backup socket", zap.String("path", socketPath))
	dialer := net.Dialer{Timeout: backupDialTimeout}
	conn, err := dialer.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if unixConn, ok := conn.(*net.UnixConn); ok {
		if err := unixConn.SetDeadline(time.Now().Add(backupStreamTimeout)); err != nil {
			return err
		}
	}

	zap.L().Debug("Writing snapshot to socket")
	return writeCombinedSnapshot(conn, db, connDB)
}

func exportBackupFile(conf *config.Configuration, writer io.Writer) error {
	backupPath := conf.DB.Backup.Path
	if backupPath == "" {
		return errors.New("backup path is empty")
	}

	zap.L().Debug("Opening backup file for export", zap.String("path", backupPath))
	stat, err := os.Stat(backupPath)
	if err != nil {
		return errors.Wrap(err, "failed to stat backup file")
	}
	if stat.IsDir() {
		return errors.New("backup path is a directory")
	}

	file, err := os.Open(backupPath)
	if err != nil {
		return errors.Wrap(err, "failed to open backup file")
	}
	defer file.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return errors.Wrap(err, "failed to stream backup file")
	}

	return nil
}

func writeCombinedSnapshotFile(db *database.Database, connDB *connections.Connections, backupPath string) (err error) {
	if backupPath == "" {
		return errors.New("backup path is empty")
	}

	tmpDir := filepath.Dir(backupPath)
	tmpFile, err := os.CreateTemp(tmpDir, "trakx-db-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp backup file")
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	if err = writeCombinedSnapshot(tmpFile, db, connDB); err != nil {
		return errors.Wrap(err, "failed to write combined snapshot")
	}
	if err = tmpFile.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync backup file")
	}
	if err = tmpFile.Close(); err != nil {
		return errors.Wrap(err, "failed to close backup file")
	}
	if err = os.Rename(tmpPath, backupPath); err != nil {
		return errors.Wrap(err, "failed to replace backup file")
	}
	if err = os.Chmod(backupPath, backupFilePermissions); err != nil {
		return errors.Wrap(err, "failed to set backup file permissions")
	}

	return nil
}

func readProcessID(path string) (int, error) {
	zap.L().Debug("Reading process id file", zap.String("path", path))
	contents, err := os.ReadFile(path)
	if err != nil {
		return -1, err
	}
	if len(contents) == 0 {
		return -1, errors.New("process id file is empty")
	}

	processID, err := strconv.Atoi(string(contents))
	if err != nil {
		return -1, errors.New("failed to parse process id file")
	}

	return processID, nil
}

func isProcessAlive(processID int) bool {
	if processID <= 0 {
		return false
	}
	zap.L().Debug("Checking process liveness", zap.Int("pid", processID))
	err := syscall.Kill(processID, syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func backupSocketPath(processID int) string {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d.sock", backupSocketPrefix, processID))
	zap.L().Debug("Resolved backup socket path", zap.Int("pid", processID), zap.String("path", path))
	return path
}

func removeSocketPath(path string) error {
	zap.L().Debug("Removing backup socket path", zap.String("path", path))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove backup socket path")
	}
	return nil
}

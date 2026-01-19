package backup

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	filePermissions = 0640 // rw-r-----
)

// FileSource reads backups from a file.
type FileSource struct {
	Path string
}

// NewFileSource creates a new file source.
func NewFileSource(path string) *FileSource {
	return &FileSource{Path: path}
}

func (s *FileSource) Open() (io.ReadCloser, error) {
	if s.Path == "" {
		return nil, errors.New("backup file path is empty")
	}

	stat, err := os.Stat(s.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to stat backup file")
	}
	if stat.IsDir() {
		return nil, errors.New("backup path is a directory")
	}

	file, err := os.Open(s.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open backup file")
	}

	return file, nil
}

func (s *FileSource) Description() string {
	return "file:" + s.Path
}

// FileDestination writes backups to a file atomically (temp file + rename).
type FileDestination struct {
	Path string
}

// NewFileDestination creates a new file destination.
func NewFileDestination(path string) *FileDestination {
	return &FileDestination{Path: path}
}

func (d *FileDestination) Create() (io.WriteCloser, error) {
	if d.Path == "" {
		return nil, errors.New("backup file path is empty")
	}

	tmpDir := filepath.Dir(d.Path)
	tmpFile, err := os.CreateTemp(tmpDir, "trakx-backup-*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp backup file")
	}

	// Return a wrapper that handles atomic rename on close
	return &atomicFileWriter{
		file:      tmpFile,
		tmpPath:   tmpFile.Name(),
		finalPath: d.Path,
	}, nil
}

func (d *FileDestination) Description() string {
	return "file:" + d.Path
}

// atomicFileWriter wraps a temp file and atomically renames it on successful close.
type atomicFileWriter struct {
	file      *os.File
	tmpPath   string
	finalPath string
	closed    bool
	failed    bool
}

func (w *atomicFileWriter) Write(p []byte) (n int, err error) {
	if w.failed {
		return 0, errors.New("writer is in failed state")
	}
	n, err = w.file.Write(p)
	if err != nil {
		w.failed = true
		w.cleanup()
	}
	return n, err
}

func (w *atomicFileWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true

	if w.failed {
		return w.cleanup()
	}

	if err := w.file.Sync(); err != nil {
		w.failed = true
		w.cleanup()
		return errors.Wrap(err, "failed to sync backup file")
	}

	if err := w.file.Close(); err != nil {
		w.failed = true
		w.cleanup()
		return errors.Wrap(err, "failed to close backup file")
	}

	if err := os.Rename(w.tmpPath, w.finalPath); err != nil {
		w.failed = true
		w.cleanup()
		return errors.Wrap(err, "failed to rename backup file")
	}

	if err := os.Chmod(w.finalPath, filePermissions); err != nil {
		return errors.Wrap(err, "failed to set backup file permissions")
	}

	return nil
}

func (w *atomicFileWriter) cleanup() error {
	w.file.Close()
	return os.Remove(w.tmpPath)
}

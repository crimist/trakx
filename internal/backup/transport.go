package backup

import (
	"io"

	"github.com/crimist/trakx/internal/storage"
)

// Source represents a source from which backups can be read.
// Examples: file, socket from daemon, stdin stream
type Source interface {
	// Open prepares the source for reading and returns a reader.
	// The returned closer must be called when done, even if reader wasn't fully consumed.
	Open() (r io.ReadCloser, err error)

	// Description returns a human-readable description of this source
	// (e.g., "file:/path/to/backup", "socket:pid=1234", "stream:stdin")
	Description() string
}

// Destination represents a destination to which backups can be written.
// Examples: file (with atomic rename), stdout stream
type Destination interface {
	// Create prepares the destination for writing and returns a writer.
	// Implementations should handle atomicity (e.g., temp file + rename).
	// The returned closer must be called to finalize the write.
	Create() (w io.WriteCloser, err error)

	// Description returns a human-readable description of this destination
	// (e.g., "file:/path/to/backup", "stream:stdout")
	Description() string
}

// SnapshotProvider provides access to components that can be snapshotted.
// This decouples the backup system from concrete database types.
type SnapshotProvider interface {
	// GetSnapshottables returns the components to include in the backup.
	// Returns (database, connections) - either can be nil.
	GetSnapshottables() (db storage.Snapshotter, connDB storage.Snapshotter)
}

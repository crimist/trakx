package backup

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/crimist/trakx/internal/storage"
)

/*
Combined snapshot format:

The combined snapshot contains both the connections database and the torrent database.

Format:
  [MAGIC:9][VERSION:2][CONN_LEN:8][CONN_DATA:CONN_LEN][DATABASE]

Where:
  - MAGIC is "TRAKXSNAP"
  - VERSION is a uint16 version number (currently 1)
  - CONN_LEN is the length of the connections snapshot data (uint64)
  - CONN_DATA is the connections snapshot (in its own internal format)
  - DATABASE is the database snapshot (with its own header and format)

The connections data is length-prefixed to allow it to be extracted separately,
while the database streams directly without length prefixing for efficiency.
*/

const (
	combinedSnapshotMagic   = "TRAKXSNAP"
	combinedSnapshotVersion = uint16(1)
)

var (
	errCombinedSnapshotMagic   = errors.New("invalid combined snapshot magic")
	errCombinedSnapshotVersion = errors.New("unsupported combined snapshot version")
)

// WriteCombined writes a combined snapshot containing both database and connections to the writer.
// Both components are optional - pass nil to skip writing that component.
func WriteCombined(writer io.Writer, db, connDB storage.Snapshotter) error {
	bufWriter := bufio.NewWriter(writer)

	if _, err := bufWriter.WriteString(combinedSnapshotMagic); err != nil {
		return err
	}
	if err := binary.Write(bufWriter, binary.LittleEndian, combinedSnapshotVersion); err != nil {
		return err
	}

	var connBytes []byte
	if connDB != nil {
		var connBuf bytes.Buffer
		if err := connDB.Snapshot(&connBuf); err != nil {
			return err
		}
		connBytes = connBuf.Bytes()
	}

	if err := binary.Write(bufWriter, binary.LittleEndian, uint64(len(connBytes))); err != nil {
		return err
	}
	if len(connBytes) > 0 {
		if _, err := bufWriter.Write(connBytes); err != nil {
			return err
		}
	}

	if err := bufWriter.Flush(); err != nil {
		return err
	}

	if db != nil {
		if err := db.Snapshot(writer); err != nil {
			return err
		}
	}

	return nil
}

// RestoreCombined reads a combined snapshot and returns the connections data and a reader for the database.
// The connections are read into memory (as they're typically small and needed later in startup).
// The database reader should be used immediately or closed to avoid resource leaks.
func RestoreCombined(reader io.Reader) (connData []byte, dbReader io.ReadCloser, err error) {
	bufReader := bufio.NewReader(reader)

	magic := make([]byte, len(combinedSnapshotMagic))
	if _, err = io.ReadFull(bufReader, magic); err != nil {
		return nil, nil, err
	}
	if string(magic) != combinedSnapshotMagic {
		return nil, nil, errCombinedSnapshotMagic
	}

	var version uint16
	if err = binary.Read(bufReader, binary.LittleEndian, &version); err != nil {
		return nil, nil, err
	}
	if version != combinedSnapshotVersion {
		return nil, nil, errCombinedSnapshotVersion
	}

	var connLen uint64
	if err = binary.Read(bufReader, binary.LittleEndian, &connLen); err != nil {
		return nil, nil, err
	}

	if connLen > 0 {
		connData = make([]byte, connLen)
		if _, err = io.ReadFull(bufReader, connData); err != nil {
			return nil, nil, err
		}
	}

	wrapped := &snapshotReader{Reader: bufReader}
	if closer, ok := reader.(io.Closer); ok {
		wrapped.closer = closer
	}

	return connData, wrapped, nil
}

// snapshotReader wraps a bufio.Reader and preserves the underlying closer
type snapshotReader struct {
	*bufio.Reader
	closer io.Closer
}

func (r *snapshotReader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

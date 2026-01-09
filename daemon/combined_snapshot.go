package daemon

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"

	"github.com/crimist/trakx/storage/database"
	"github.com/crimist/trakx/tracker/udp/connections"
)

const (
	combinedSnapshotMagic       = "TRAKXSNAP"
	combinedSnapshotVersion     = uint16(1)
	combinedSnapshotMagicSize   = len(combinedSnapshotMagic)
)

var (
	errCombinedSnapshotMagic   = errors.New("invalid combined snapshot magic")
	errCombinedSnapshotVersion = errors.New("unsupported combined snapshot version")
	errCombinedSnapshotSize    = errors.New("connections snapshot too large")
)

type snapshotReader struct {
	*bufio.Reader
	closer io.Closer
}

func (reader *snapshotReader) Close() error {
	if reader.closer != nil {
		return reader.closer.Close()
	}
	return nil
}

func splitCombinedSnapshot(reader io.Reader) ([]byte, io.Reader, error) {
	bufReader := bufio.NewReader(reader)

	header := make([]byte, combinedSnapshotMagicSize)
	if _, err := io.ReadFull(bufReader, header); err != nil {
		return nil, nil, err
	}
	if string(header) != combinedSnapshotMagic {
		return nil, nil, errCombinedSnapshotMagic
	}

	var version uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &version); err != nil {
		return nil, nil, err
	}
	if version != combinedSnapshotVersion {
		return nil, nil, errCombinedSnapshotVersion
	}

	var connLen uint64
	if err := binary.Read(bufReader, binary.LittleEndian, &connLen); err != nil {
		return nil, nil, err
	}

	maxInt := int(^uint(0) >> 1)
	if connLen > uint64(maxInt) {
		return nil, nil, errCombinedSnapshotSize
	}

	var connBytes []byte
	if connLen > 0 {
		connBytes = make([]byte, int(connLen))
		if _, err := io.ReadFull(bufReader, connBytes); err != nil {
			return nil, nil, err
		}
	}

	wrapped := &snapshotReader{Reader: bufReader}
	if closer, ok := reader.(io.Closer); ok {
		wrapped.closer = closer
	}

	return connBytes, wrapped, nil
}

func writeCombinedSnapshot(writer io.Writer, db *database.Database, connDB *connections.Connections) error {
	bufWriter := bufio.NewWriter(writer)

	if _, err := bufWriter.WriteString(combinedSnapshotMagic); err != nil {
		return err
	}
	if err := binary.Write(bufWriter, binary.LittleEndian, combinedSnapshotVersion); err != nil {
		return err
	}

	var connBytes []byte
	if connDB != nil {
		var err error
		connBytes, err = connDB.Marshal()
		if err != nil {
			return err
		}
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

	return db.Snapshot(writer)
}

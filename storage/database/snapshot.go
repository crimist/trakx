package database

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/crimist/trakx/storage"
)

const (
	snapshotMagic       = "TRAKXDB"
	snapshotVersion     = uint16(1)
	snapshotMagicSize   = len(snapshotMagic)
	snapshotVersionSize = 2 // uint16, little endian
	snapshotHeaderSize  = snapshotMagicSize + snapshotVersionSize
)

var errSnapshotVersion = errors.New("unsupported snapshot version")

// Snapshot writes a full database snapshot to the provided writer.
func (db *Database) Snapshot(writer io.Writer) error {
	bufWriter := bufio.NewWriter(writer)
	if err := writeSnapshotHeader(bufWriter); err != nil {
		return err
	}
	if err := encodeBinaryToWriter(db, bufWriter); err != nil {
		return err
	}

	return bufWriter.Flush()
}

// Restore loads a database snapshot from the provided reader.
// It also supports the legacy headerless binary format.
func (db *Database) Restore(reader io.Reader) error {
	bufReader := bufio.NewReader(reader)
	header := make([]byte, snapshotMagicSize)
	if _, err := io.ReadFull(bufReader, header); err != nil {
		return err
	}

	if string(header) != snapshotMagic {
		legacyReader := io.MultiReader(bytes.NewReader(header), bufReader)
		return db.restoreBinary(legacyReader)
	}

	var version uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &version); err != nil {
		return err
	}
	if version != snapshotVersion {
		return errSnapshotVersion
	}

	return db.restoreBinary(bufReader)
}

func writeSnapshotHeader(writer *bufio.Writer) error {
	if _, err := writer.WriteString(snapshotMagic); err != nil {
		return err
	}

	return binary.Write(writer, binary.LittleEndian, snapshotVersion)
}

func (db *Database) restoreBinary(reader io.Reader) error {
	db.mutex.Lock()
	db.torrents = make(map[storage.Hash]*Torrent, 1)
	db.mutex.Unlock()

	_, _, err := decodeBinaryFromReader(db, reader)
	return err
}

package connections

/*
Binary encoding for connection snapshots.

This encoder is optimized for speed and simplicity. Connection snapshots are typically
small (embedded in combined snapshots) and are buffered into memory.

Format:
  [count:u32][entries...]

Where each entry is:
  [flags:u8][ip:4/16][port:u16][id:u64][timestamp:i64]

Flags byte layout:
  - Bit 0: IP version (0=IPv6, 1=IPv4)
  - Bits 1-7: Reserved for future use

Entry sizes:
  - IPv4 entry: 1 + 4 + 2 + 8 + 8 = 23 bytes
  - IPv6 entry: 1 + 16 + 2 + 8 + 8 = 35 bytes

The format uses uint32 for count (max 4.2B connections) and keeps int64 timestamps
for future-proofing beyond year 2106.
*/

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net/netip"
)

const (
	// Flag byte layout
	flagIPv4 = 1 << 0 // Bit 0: IP version (0=IPv6, 1=IPv4)

	// Maximum entry size: flags(1) + IPv6(16) + port(2) + ID(8) + timestamp(8) = 35 bytes
	entryMaxSize = 35
)

// Snapshot writes a full connections snapshot to the provided writer using binary encoding.
func (connections *Connections) Snapshot(writer io.Writer) error {
	bufWriter, ok := writer.(*bufio.Writer)
	if !ok {
		bufWriter = bufio.NewWriter(writer)
		defer bufWriter.Flush()
	}

	// Copy associations under lock, then release to avoid blocking Create() calls
	connections.mutex.RLock()
	associationsCopy := make(associations, len(connections.associations))
	for addr, entry := range connections.associations {
		associationsCopy[addr] = entry
	}
	connections.mutex.RUnlock()

	// Write entry count
	if err := binary.Write(bufWriter, binary.LittleEndian, uint32(len(associationsCopy))); err != nil {
		return err
	}

	// Pre-allocate scratch buffer for entries
	scratch := make([]byte, entryMaxSize)

	for addr, entry := range associationsCopy {
		offset := 0

		// Flags byte
		flags := byte(0)
		if addr.Addr().Is4() {
			flags |= flagIPv4
		}
		scratch[offset] = flags
		offset++

		// IP address
		if addr.Addr().Is4() {
			ip4 := addr.Addr().As4()
			copy(scratch[offset:offset+4], ip4[:])
			offset += 4
		} else {
			ip6 := addr.Addr().As16()
			copy(scratch[offset:offset+16], ip6[:])
			offset += 16
		}

		// Port
		binary.LittleEndian.PutUint16(scratch[offset:offset+2], addr.Port())
		offset += 2

		// ID (uint64)
		binary.LittleEndian.PutUint64(scratch[offset:offset+8], entry.ID)
		offset += 8

		// Timestamp (int64)
		binary.LittleEndian.PutUint64(scratch[offset:offset+8], uint64(entry.TimeStamp))
		offset += 8

		if _, err := bufWriter.Write(scratch[:offset]); err != nil {
			return err
		}
	}

	return nil
}

// Restore loads a connections snapshot from the provided reader.
func (connections *Connections) Restore(reader io.Reader) error {
	bufReader, ok := reader.(*bufio.Reader)
	if !ok {
		bufReader = bufio.NewReader(reader)
	}

	// Read entry count
	var count uint32
	if err := binary.Read(bufReader, binary.LittleEndian, &count); err != nil {
		return err
	}

	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	connections.associations = make(associations, count)
	scratch := make([]byte, entryMaxSize)

	for i := uint32(0); i < count; i++ {
		if _, err := io.ReadFull(bufReader, scratch[:1]); err != nil {
			return err
		}
		flags := scratch[0]
		isIPv4 := (flags & flagIPv4) != 0

		var addr netip.Addr
		var ipSize int
		if isIPv4 {
			ipSize = 4
		} else {
			ipSize = 16
		}

		readSize := ipSize + 2 + 8 + 8
		if _, err := io.ReadFull(bufReader, scratch[:readSize]); err != nil {
			return err
		}

		offset := 0
		if isIPv4 {
			addr = netip.AddrFrom4([4]byte(scratch[offset : offset+4]))
			offset += 4
		} else {
			addr = netip.AddrFrom16([16]byte(scratch[offset : offset+16]))
			offset += 16
		}

		port := binary.LittleEndian.Uint16(scratch[offset : offset+2])
		offset += 2

		id := binary.LittleEndian.Uint64(scratch[offset : offset+8])
		offset += 8

		timestamp := int64(binary.LittleEndian.Uint64(scratch[offset : offset+8]))

		addrPort := netip.AddrPortFrom(addr, port)
		connections.associations[addrPort] = associationEntry{
			ID:        id,
			TimeStamp: timestamp,
		}
	}

	return nil
}

// Marshal is a convenience method that returns the snapshot as a byte slice.
// For large snapshots, prefer using Snapshot() with a writer directly.
func (connections *Connections) Marshal() ([]byte, error) {
	var buffer bytes.Buffer
	if err := connections.Snapshot(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Unmarshal is a convenience method that restores from a byte slice.
// For large snapshots, prefer using Restore() with a reader directly.
func (connections *Connections) Unmarshal(data []byte) error {
	return connections.Restore(bytes.NewReader(data))
}

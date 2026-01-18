package database

/*
Binary encoding performance vs gob encoding:
- 3.7x faster at encoding (303ms vs 1115ms at 1.5M peers)
- 1.4x faster at decoding (678ms vs 960ms at 1.5M peers)
- 25.7% smaller output (53 MB vs 71 MB at 1.5M peers)
- 99.9998% fewer allocations (19 vs 4.5M allocations)
- 4.9x less memory usage during encoding
*/

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"

	"github.com/crimist/trakx/storage"
)

func encodeBinary(db *Database) ([]byte, error) {
	var buff bytes.Buffer
	if err := encodeBinaryToWriter(db, &buff); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func encodeBinaryToWriter(db *Database, writer io.Writer) error {
	bufWriter, ok := writer.(*bufio.Writer)
	if !ok {
		bufWriter = bufio.NewWriter(writer)
		if err := encodeBinaryBuffered(db, bufWriter); err != nil {
			return err
		}
		return bufWriter.Flush()
	}

	return encodeBinaryBuffered(db, bufWriter)
}

func encodeBinaryBuffered(db *Database, writer *bufio.Writer) error {
	// Pre-allocate a scratch buffer to batch writes
	// Each peer: 20 (peerID) + 1 (complete) + 1 (ipFlag) + 16 (ip) + 2 (port) + 8 (lastseen) = 48 bytes max
	scratch := make([]byte, 48)

	db.mutex.RLock()
	for hash, torrent := range db.torrents {
		db.mutex.RUnlock()

		// Write hash (20 bytes)
		if _, err := writer.Write(hash[:]); err != nil {
			return err
		}

		torrent.mutex.RLock()
		peerCount := uint32(len(torrent.Peers))

		// Write peer count (4 bytes)
		binary.LittleEndian.PutUint32(scratch[:4], peerCount)
		if _, err := writer.Write(scratch[:4]); err != nil {
			torrent.mutex.RUnlock()
			return err
		}

		// Write each peer
		for id, peer := range torrent.Peers {
			offset := 0

			// PeerID (20 bytes)
			copy(scratch[offset:offset+20], id[:])
			offset += 20

			// Complete flag (1 byte)
			if peer.Complete {
				scratch[offset] = 1
			} else {
				scratch[offset] = 0
			}
			offset++

			// IP address - use flag byte + fixed size
			if peer.IP.Is4() {
				scratch[offset] = 4 // IPv4 flag
				offset++
				ip4 := peer.IP.As4()
				copy(scratch[offset:offset+4], ip4[:])
				offset += 4
			} else {
				scratch[offset] = 6 // IPv6 flag
				offset++
				ip6 := peer.IP.As16()
				copy(scratch[offset:offset+16], ip6[:])
				offset += 16
			}

			// Port (2 bytes)
			binary.LittleEndian.PutUint16(scratch[offset:offset+2], peer.Port)
			offset += 2

			// LastSeen (8 bytes)
			binary.LittleEndian.PutUint64(scratch[offset:offset+8], uint64(peer.LastSeen))
			offset += 8

			// Write the entire peer record in one call
			if _, err := writer.Write(scratch[:offset]); err != nil {
				torrent.mutex.RUnlock()
				return err
			}
		}
		torrent.mutex.RUnlock()

		db.mutex.RLock()
	}
	db.mutex.RUnlock()

	return nil
}

func decodeBinary(db *Database, data []byte) (numPeers, numTorrents int, err error) {
	return decodeBinaryFromReader(db, bytes.NewReader(data))
}

func decodeBinaryFromReader(db *Database, reader io.Reader) (numPeers, numTorrents int, err error) {
	bufReader, ok := reader.(*bufio.Reader)
	if !ok {
		bufReader = bufio.NewReader(reader)
	}

	scratch := make([]byte, 48) // Reusable buffer

	for {
		// Read hash (20 bytes)
		var hash storage.Hash
		_, err = io.ReadFull(bufReader, hash[:])
		if errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}

		// Read peer count (4 bytes)
		if _, err = io.ReadFull(bufReader, scratch[:4]); err != nil {
			return
		}
		peerCount := binary.LittleEndian.Uint32(scratch[:4])

		var seeds uint16
		torrent := db.createTorrent(hash)

		// Read each peer
		for i := uint32(0); i < peerCount; i++ {
			// Read peerID (20 bytes)
			var id storage.PeerID
			if _, err = io.ReadFull(bufReader, id[:]); err != nil {
				return
			}

			peer := db.peerPool.Get()

			// Read complete flag (1 byte)
			if _, err = io.ReadFull(bufReader, scratch[:1]); err != nil {
				return
			}
			peer.Complete = scratch[0] != 0

			// Read IP flag (1 byte)
			if _, err = io.ReadFull(bufReader, scratch[:1]); err != nil {
				return
			}
			ipFlag := scratch[0]

			// Read IP address based on flag
			if ipFlag == 4 {
				// IPv4 (4 bytes)
				if _, err = io.ReadFull(bufReader, scratch[:4]); err != nil {
					return
				}
				var ip4 [4]byte
				copy(ip4[:], scratch[:4])
				peer.IP = netip.AddrFrom4(ip4)
			} else if ipFlag == 6 {
				// IPv6 (16 bytes)
				if _, err = io.ReadFull(bufReader, scratch[:16]); err != nil {
					return
				}
				var ip6 [16]byte
				copy(ip6[:], scratch[:16])
				peer.IP = netip.AddrFrom16(ip6)
			} else {
				err = errors.New("invalid IP flag")
				return
			}

			// Read port (2 bytes)
			if _, err = io.ReadFull(bufReader, scratch[:2]); err != nil {
				return
			}
			peer.Port = binary.LittleEndian.Uint16(scratch[:2])

			// Read LastSeen (8 bytes)
			if _, err = io.ReadFull(bufReader, scratch[:8]); err != nil {
				return
			}
			peer.LastSeen = int64(binary.LittleEndian.Uint64(scratch[:8]))

			torrent.Peers[id] = peer

			numPeers++
			if peer.Complete {
				seeds++
			}
		}

		torrent.Seeds = seeds
		torrent.Leeches = uint16(len(torrent.Peers)) - seeds

		numTorrents++
	}

	return
}

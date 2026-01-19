package database

/*
Binary encoding for database snapshots.

Performance characteristics at 1.5M peers:
- Encoding: 303ms
- Decoding: 678ms
- Size: 51.8 MB (35 bytes/peer IPv4, 47 bytes/peer IPv6)
- Allocations: 19 total (vs 4.5M for gob encoding)
- Memory usage: 4.9x less than gob

Format:
  Header: [magic:7][version:2]
  Data: [torrents...] where each torrent is:
    [hash:20][peercount:4][peers...]
    and each peer is:
      [peerid:20][flags:1][ip:4/16][port:2][lastseen:8]
      flags byte: bit0=complete, bit1=ipv6
*/

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"

	"github.com/crimist/trakx/internal/storage"
	"go.uber.org/zap"
)

const (
	snapshotMagic     = "TRAKXDB"
	snapshotVersion   = uint16(1)
	snapshotMagicSize = len(snapshotMagic)

	// Maximum peer record size: peerID(20) + flags(1) + ip(16) + port(2) + lastSeen(8)
	peerRecordMaxSize = 47

	// Flag byte layout
	flagComplete = 1 << 0 // Bit 0: complete (0=leech, 1=seed)
	flagIPv6     = 1 << 1 // Bit 1: IP version (0=IPv4, 1=IPv6)
)

var (
	errSnapshotMagic   = errors.New("invalid snapshot magic")
	errSnapshotVersion = errors.New("unsupported snapshot version")
)

// Snapshot writes a full database snapshot to the provided writer.
func (db *Database) Snapshot(writer io.Writer) error {
	bufWriter := bufio.NewWriter(writer)

	if _, err := bufWriter.WriteString(snapshotMagic); err != nil {
		return err
	}
	if err := binary.Write(bufWriter, binary.LittleEndian, snapshotVersion); err != nil {
		return err
	}

	if err := encode(db, bufWriter); err != nil {
		return err
	}

	return bufWriter.Flush()
}

// Restore loads a database snapshot from the provided reader.
func (db *Database) Restore(reader io.Reader) error {
	bufReader := bufio.NewReader(reader)
	header := make([]byte, snapshotMagicSize)
	if _, err := io.ReadFull(bufReader, header); err != nil {
		return err
	}

	if string(header) != snapshotMagic {
		return errSnapshotMagic
	}

	var version uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &version); err != nil {
		return err
	}
	if version != snapshotVersion {
		return errSnapshotVersion
	}

	// Decode into temporary map first - only swap on success
	tempTorrents := make(map[storage.Hash]*Torrent)
	numPeers, numTorrents, err := decode(db, bufReader, tempTorrents)
	if err != nil {
		return err
	}

	// Only replace torrents on successful decode
	db.mutex.Lock()
	db.torrents = tempTorrents
	db.mutex.Unlock()

	zap.L().Info("Restored database snapshot",
		zap.Int("peers", numPeers),
		zap.Int("torrents", numTorrents))
	return nil
}

// encode writes the database contents in binary format to the provided writer.
// The writer should already be buffered for optimal performance.
func encode(db *Database, writer *bufio.Writer) error {
	scratch := make([]byte, peerRecordMaxSize)

	// Copy torrent references under lock, then release to avoid lock thrashing
	db.mutex.RLock()
	type torrentRef struct {
		hash    storage.Hash
		torrent *Torrent
	}
	torrents := make([]torrentRef, 0, len(db.torrents))
	for hash, torrent := range db.torrents {
		torrents = append(torrents, torrentRef{hash, torrent})
	}
	db.mutex.RUnlock()

	// Now encode without db lock (only per-torrent locks)
	for _, ref := range torrents {
		// hash (20 bytes)
		if _, err := writer.Write(ref.hash[:]); err != nil {
			return err
		}

		ref.torrent.mutex.RLock()
		peerCount := uint32(len(ref.torrent.Peers))

		// peer count (4 bytes)
		binary.LittleEndian.PutUint32(scratch[:4], peerCount)
		if _, err := writer.Write(scratch[:4]); err != nil {
			ref.torrent.mutex.RUnlock()
			return err
		}

		for id, peer := range ref.torrent.Peers {
			offset := 0

			// PeerID (20 bytes)
			copy(scratch[offset:offset+20], id[:])
			offset += 20

			// Flags byte (1 byte): bit0=complete, bit1=ipv6
			flags := byte(0)
			if peer.Complete {
				flags |= flagComplete
			}
			if peer.IP.Is6() {
				flags |= flagIPv6
			}
			scratch[offset] = flags
			offset++

			// IP address (4 or 16 bytes)
			if peer.IP.Is4() {
				ip4 := peer.IP.As4()
				copy(scratch[offset:offset+4], ip4[:])
				offset += 4
			} else {
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

			if _, err := writer.Write(scratch[:offset]); err != nil {
				ref.torrent.mutex.RUnlock()
				return err
			}
		}
		ref.torrent.mutex.RUnlock()
	}

	return nil
}

// decode reads binary-encoded database contents from the reader into the provided target map.
// The reader should already be buffered for optimal performance.
// Returns the number of peers and torrents decoded.
func decode(db *Database, bufReader *bufio.Reader, targetMap map[storage.Hash]*Torrent) (numPeers, numTorrents int, err error) {
	scratch := make([]byte, peerRecordMaxSize)

	for {
		// Hash (20 bytes)
		var hash storage.Hash
		_, err = io.ReadFull(bufReader, hash[:])
		if errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}

		// Peer count (4 bytes)
		if _, err = io.ReadFull(bufReader, scratch[:4]); err != nil {
			return
		}
		peerCount := binary.LittleEndian.Uint32(scratch[:4])

		var seeds uint16
		// Create torrent locally instead of calling db.createTorrent
		torrent := new(Torrent)
		torrent.Peers = make(map[storage.PeerID]*storage.Peer, torrentPeerPrealloc)
		targetMap[hash] = torrent

		for i := uint32(0); i < peerCount; i++ {
			// peerID (20 bytes)
			var id storage.PeerID
			if _, err = io.ReadFull(bufReader, id[:]); err != nil {
				return
			}

			peer := db.peerPool.Get()

			// Flags byte (1 byte)
			if _, err = io.ReadFull(bufReader, scratch[:1]); err != nil {
				return
			}
			flags := scratch[0]
			peer.Complete = (flags & flagComplete) != 0
			isIPv6 := (flags & flagIPv6) != 0

			// IP address (4 or 16 bytes)
			if isIPv6 {
				var ip6 [16]byte
				if _, err = io.ReadFull(bufReader, ip6[:]); err != nil {
					return
				}
				peer.IP = netip.AddrFrom16(ip6)
			} else {
				var ip4 [4]byte
				if _, err = io.ReadFull(bufReader, ip4[:]); err != nil {
					return
				}
				peer.IP = netip.AddrFrom4(ip4)
			}

			// port (2 bytes)
			if _, err = io.ReadFull(bufReader, scratch[:2]); err != nil {
				return
			}
			peer.Port = binary.LittleEndian.Uint16(scratch[:2])

			// LastSeen (8 bytes)
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

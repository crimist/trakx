package database

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"

	"github.com/crimist/trakx/storage"
)

// binary coders are better than gob coders below ~1.5 million peers

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
	db.mutex.RLock()
	for hash, torrent := range db.torrents {
		db.mutex.RUnlock()

		// hash + number of torrent peers
		if err := binary.Write(writer, binary.LittleEndian, &hash); err != nil {
			return err
		}
		torrent.mutex.RLock()
		if err := binary.Write(writer, binary.LittleEndian, uint32(len(torrent.Peers))); err != nil {
			torrent.mutex.RUnlock()
			return err
		}

		// peerid + peer
		for id, peer := range torrent.Peers {
			if err := binary.Write(writer, binary.LittleEndian, &id); err != nil {
				torrent.mutex.RUnlock()
				return err
			}

			addrSlice := peer.IP.AsSlice()
			if err := binary.Write(writer, binary.LittleEndian, peer.Complete); err != nil {
				torrent.mutex.RUnlock()
				return err
			}
			if err := binary.Write(writer, binary.LittleEndian, int32(len(addrSlice))); err != nil {
				torrent.mutex.RUnlock()
				return err
			}
			if err := binary.Write(writer, binary.LittleEndian, addrSlice); err != nil {
				torrent.mutex.RUnlock()
				return err
			}
			if err := binary.Write(writer, binary.LittleEndian, peer.Port); err != nil {
				torrent.mutex.RUnlock()
				return err
			}
			if err := binary.Write(writer, binary.LittleEndian, peer.LastSeen); err != nil {
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

	for {
		// decode hash + number of torrent peers
		var hash storage.Hash
		err = binary.Read(bufReader, binary.LittleEndian, &hash)
		if errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}

		var peerCount uint32
		var seeds uint16
		torrent := db.createTorrent(hash)
		if err = binary.Read(bufReader, binary.LittleEndian, &peerCount); err != nil {
			return
		}

		// decode peerid and peers
		for ; peerCount > 0; peerCount-- {
			var id storage.PeerID
			if err = binary.Read(bufReader, binary.LittleEndian, &id); err != nil {
				return
			}

			peer := db.peerPool.Get()
			var addrSliceLen int32
			if err = binary.Read(bufReader, binary.LittleEndian, &peer.Complete); err != nil {
				return
			}
			if err = binary.Read(bufReader, binary.LittleEndian, &addrSliceLen); err != nil {
				return
			}
			addrSlice := make([]byte, addrSliceLen)
			if err = binary.Read(bufReader, binary.LittleEndian, &addrSlice); err != nil {
				return
			}
			ip, ok := netip.AddrFromSlice(addrSlice)
			if !ok {
				err = errors.New("AddrFromSlice failed")
				return
			}
			peer.IP = ip
			if err = binary.Read(bufReader, binary.LittleEndian, &peer.Port); err != nil {
				return
			}
			if err = binary.Read(bufReader, binary.LittleEndian, &peer.LastSeen); err != nil {
				return
			}
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

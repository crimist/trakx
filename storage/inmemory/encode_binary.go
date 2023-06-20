package inmemory

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

func encodeBinary(db *InMemory) ([]byte, error) {
	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)

	db.mutex.RLock()
	for hash, torrent := range db.torrents {
		db.mutex.RUnlock()

		// hash + number of torrent peers
		if err := binary.Write(writer, binary.LittleEndian, &hash); err != nil {
			return nil, err
		}
		torrent.mutex.RLock()
		if err := binary.Write(writer, binary.LittleEndian, uint32(len(torrent.Peers))); err != nil {
			torrent.mutex.RUnlock()
			return nil, err
		}

		// peerid + peer
		for id, peer := range torrent.Peers {
			if err := binary.Write(writer, binary.LittleEndian, &id); err != nil {
				torrent.mutex.RUnlock()
				return nil, err
			}

			addrSlice := peer.IP.AsSlice()
			if err := binary.Write(writer, binary.LittleEndian, peer.Complete); err != nil {
				torrent.mutex.RUnlock()
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, int32(len(addrSlice))); err != nil {
				torrent.mutex.RUnlock()
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, addrSlice); err != nil {
				torrent.mutex.RUnlock()
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, peer.Port); err != nil {
				torrent.mutex.RUnlock()
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, peer.LastSeen); err != nil {
				torrent.mutex.RUnlock()
				return nil, err
			}
		}
		torrent.mutex.RUnlock()

		db.mutex.RLock()
	}
	db.mutex.RUnlock()

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func decodeBinary(db *InMemory, data []byte) (numPeers, numTorrents int, err error) {
	reader := bufio.NewReader(bytes.NewBuffer(data))

	for {
		// decode hash + number of torrent peers
		var hash storage.Hash
		err = binary.Read(reader, binary.LittleEndian, &hash)
		if errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}

		var peerCount uint32
		var seeds uint16
		torrent := db.createTorrent(hash)
		if err = binary.Read(reader, binary.LittleEndian, &peerCount); err != nil {
			return
		}

		// decode peerid and peers
		for ; peerCount > 0; peerCount-- {
			var id storage.PeerID
			if err = binary.Read(reader, binary.LittleEndian, &id); err != nil {
				return
			}

			peer := db.peerPool.Get()
			var addrSliceLen int32
			if err = binary.Read(reader, binary.LittleEndian, &peer.Complete); err != nil {
				return
			}
			if err = binary.Read(reader, binary.LittleEndian, &addrSliceLen); err != nil {
				return
			}
			addrSlice := make([]byte, addrSliceLen)
			if err = binary.Read(reader, binary.LittleEndian, &addrSlice); err != nil {
				return
			}
			ip, ok := netip.AddrFromSlice(addrSlice)
			if !ok {
				err = errors.New("AddrFromSlice failed")
				return
			}
			peer.IP = ip
			if err = binary.Read(reader, binary.LittleEndian, &peer.Port); err != nil {
				return
			}
			if err = binary.Read(reader, binary.LittleEndian, &peer.LastSeen); err != nil {
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

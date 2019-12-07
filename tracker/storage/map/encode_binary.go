package gomap

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/crimist/trakx/tracker/storage"
)

func (db *Memory) encodeBinary() ([]byte, error) {
	var buff bytes.Buffer
	w := bufio.NewWriter(&buff)

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		binary.Write(w, binary.LittleEndian, hash)
		binary.Write(w, binary.LittleEndian, uint32(len(submap.peers)))

		submap.RLock()
		for id, peer := range submap.peers {
			binary.Write(w, binary.LittleEndian, id)
			binary.Write(w, binary.LittleEndian, peer)
		}
		submap.RUnlock()

		db.mu.RLock()
	}
	db.mu.RUnlock()

	if err := w.Flush(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (db *Memory) decodeBinary(data []byte) (peers, hashes int, err error) {
	db.make()

	buff := bytes.NewBuffer(data)
	w := bufio.NewReader(buff)

	for {
		var hash storage.Hash
		err = binary.Read(w, binary.LittleEndian, &hash)
		if errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}

		peermap := db.makePeermap(&hash)

		var count uint32
		if err = binary.Read(w, binary.LittleEndian, &count); err != nil {
			return
		}

		for ; count > 0; count-- {
			var id storage.PeerID
			if err = binary.Read(w, binary.LittleEndian, &id); err != nil {
				return
			}
			var peer storage.Peer
			if err = binary.Read(w, binary.LittleEndian, &peer); err != nil {
				return
			}

			peermap.peers[id] = &peer
			peers++
		}

		hashes++
	}

	return
}

package gomap

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"io"

	"github.com/crimist/trakx/tracker/storage"
)

func (db *Memory) encodeGob() ([]byte, error) {
	var buff bytes.Buffer
	w := bufio.NewWriter(&buff)
	encoder := gob.NewEncoder(w)

	db.mutex.RLock()
	for hash, submap := range db.hashmap {
		if err := encoder.Encode(hash); err != nil {
			return nil, err
		}
		if err := encoder.Encode(len(submap.peers)); err != nil {
			return nil, err
		}

		submap.RLock()
		for id, peer := range submap.peers {
			if err := encoder.Encode(id); err != nil {
				return nil, err
			}
			if err := encoder.Encode(peer); err != nil {
				return nil, err
			}
		}
		submap.RUnlock()

		db.mutex.RLock()
	}
	db.mutex.RUnlock()

	if err := w.Flush(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (db *Memory) decodeGob(data []byte) (peers int, hashes int, err error) {
	db.make()
	buff := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(bufio.NewReader(buff))

	for {
		var hash storage.Hash
		if err = decoder.Decode(&hash); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			}

			return 0, 0, err
		}

		peermap := db.makePeermap(hash)

		var count int
		if err = decoder.Decode(&count); err != nil {
			return
		}
		for ; count > 0; count-- {
			var id storage.PeerID
			if err = decoder.Decode(&id); err != nil {
				return
			}
			peer := storage.PeerChan.Get()
			if err = decoder.Decode(peer); err != nil {
				return
			}
			peermap.peers[id] = peer
			peers++
		}

		hashes++
	}

	return
}

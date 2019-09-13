package gomap

import (
	"bytes"
	"encoding/gob"

	"github.com/syc0x00/trakx/tracker/storage"
)

type encoded struct {
	Hash  storage.Hash
	Peers map[storage.PeerID]*storage.Peer
}

func (db *Memory) encode() ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	encodes := make([]encoded, db.Hashes())

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		submap.RLock()
		encodes = append(encodes, encoded{
			Hash:  hash,
			Peers: submap.peers,
		})
		submap.RUnlock()

		db.mu.RLock()
	}
	db.mu.RUnlock()

	db.mu.RLock()
	err := enc.Encode(encodes)
	db.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (db *Memory) decode(data []byte) error {
	db.make()

	var encodes []encoded
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	if err := dec.Decode(&encodes); err != nil {
		return err
	}

	for _, encd := range encodes {
		peermap := db.makePeermap(&encd.Hash)
		peermap.peers = encd.Peers
	}

	return nil
}

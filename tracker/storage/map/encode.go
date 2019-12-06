package gomap

import (
	"bytes"
	"encoding/gob"

	"github.com/crimist/trakx/tracker/storage"
)

type encoded struct {
	Hash  storage.Hash
	Peers map[storage.PeerID]*storage.Peer
}

func (db *Memory) encode() ([]byte, error) {
	var buff bytes.Buffer
	var i int

	enc := gob.NewEncoder(&buff)
	encodes := make([]encoded, db.Hashes())

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		submap.RLock()
		encodes[i] = encoded{
			Hash:  hash,
			Peers: submap.peers,
		}
		submap.RUnlock()

		i++
		db.mu.RLock()
	}
	db.mu.RUnlock()

	err := enc.Encode(encodes)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (db *Memory) decode(data []byte) (peers, hashes int, err error) {
	db.make()

	var encodes []encoded
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	if err = dec.Decode(&encodes); err != nil {
		return
	}

	for _, encd := range encodes {
		peermap := db.makePeermap(&encd.Hash)
		peermap.peers = encd.Peers

		hashes++
		peers += len(encd.Peers)
	}

	return
}

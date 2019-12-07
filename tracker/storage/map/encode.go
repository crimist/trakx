package gomap

import (
	"bytes"
	"encoding/gob"

	"github.com/crimist/trakx/tracker/storage"
)

const bytesPerPeer = 57

type peerlist struct {
	ID   *storage.PeerID
	Peer *storage.Peer
}

type encoded struct {
	Hash  storage.Hash
	Peers []peerlist
}

func (db *Memory) encode() ([]byte, error) {
	var buff bytes.Buffer
	var i, z int

	encodes := make([]encoded, db.Hashes())

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		z = 0
		plist := make([]peerlist, len(submap.peers))

		submap.RLock()
		for id, peer := range submap.peers {
			plist[z].ID = &id
			plist[z].Peer = peer
			z++
		}
		submap.RUnlock()

		encodes[i] = encoded{
			Hash:  hash,
			Peers: plist,
		}
		i++

		db.mu.RLock()
	}
	db.mu.RUnlock()

	err := gob.NewEncoder(&buff).Encode(encodes)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (db *Memory) decode(data []byte) (peers, hashes int, err error) {
	var encodes []encoded
	db.make()

	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&encodes); err != nil {
		return
	}

	for _, encoded := range encodes {
		peermap := db.makePeermap(&encoded.Hash)
		for _, peer := range encoded.Peers {
			peermap.peers[*peer.ID] = peer.Peer
			peers++
		}

		hashes++
	}

	return
}

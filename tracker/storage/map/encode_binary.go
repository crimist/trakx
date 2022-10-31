package gomap

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/crimist/trakx/tracker/storage"
)

// TODO: attempt rewrite with new storage.Peer (broken due to use of netip.Addr atm)

func (db *Memory) encodeBinary() ([]byte, error) {
	var buff bytes.Buffer
	w := bufio.NewWriter(&buff)

	db.mutex.RLock()
	for hash, submap := range db.hashmap {
		db.mutex.RUnlock()

		binary.Write(w, binary.LittleEndian, &hash)
		binary.Write(w, binary.LittleEndian, uint32(len(submap.Peers)))

		submap.mutex.RLock()
		for id, peer := range submap.Peers {
			binary.Write(w, binary.LittleEndian, &id)
			binary.Write(w, binary.LittleEndian, peer.Complete)
			ipData, _ := peer.IP.MarshalBinary()
			binary.Write(w, binary.LittleEndian, ipData)
			binary.Write(w, binary.LittleEndian, peer.Port)
			binary.Write(w, binary.LittleEndian, peer.LastSeen)
		}
		submap.mutex.RUnlock()

		db.mutex.RLock()
	}
	db.mutex.RUnlock()

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

		peermap := db.makePeermap(hash)

		var count uint32
		if err = binary.Read(w, binary.LittleEndian, &count); err != nil {
			return
		}

		for ; count > 0; count-- {
			var id storage.PeerID
			if err = binary.Read(w, binary.LittleEndian, &id); err != nil {
				return
			}
			peer := storage.PeerChan.Get()
			if err = binary.Read(w, binary.LittleEndian, peer); err != nil {
				return
			}
			peermap.Peers[id] = peer
			peers++
		}

		hashes++
	}

	return
}

package gomap

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"unsafe"

	"github.com/crimist/trakx/tracker/storage"
)

func (db *Memory) encodeBinary() ([]byte, error) {
	var buff bytes.Buffer
	w := bufio.NewWriter(&buff)

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		binary.Write(w, binary.LittleEndian, &hash)
		binary.Write(w, binary.LittleEndian, uint32(len(submap.peers)))

		submap.RLock()
		for id, peer := range submap.peers {
			binary.Write(w, binary.LittleEndian, &id)
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

// Struct to byte unsafe https://stackoverflow.com/a/56272984/6389542
const peersz = int(unsafe.Sizeof(storage.Peer{}))

func (db *Memory) encodeBinaryUnsafe() ([]byte, error) {
	var buff []byte

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		buff = append(buff, hash[:]...)

		num := make([]byte, 4)
		binary.LittleEndian.PutUint32(num, uint32(len(submap.peers)))
		buff = append(buff, num[:]...)

		submap.RLock()
		for id, peer := range submap.peers {
			buff = append(buff, id[:]...)
			buff = append(buff, (*(*[peersz]byte)(unsafe.Pointer(peer)))[:]...)
		}
		submap.RUnlock()

		db.mu.RLock()
	}
	db.mu.RUnlock()

	return buff, nil
}

// uses 5.6x less memory than it's multi alloc counterpart
func (db *Memory) encodeBinaryUnsafeAutoalloc() (buff []byte, err error) {
	defer func() {
		// recover any oob slice panics
		if x := recover(); x != nil {
			err = x.(error)
		}
	}()

	var pos int

	db.mu.Lock()
	buff = make([]byte, len(db.hashmap)*24+int(storage.Expvar.Seeds+storage.Expvar.Leeches)*36)

	for hash, submap := range db.hashmap {
		copy(buff[pos:pos+20], hash[:])
		binary.LittleEndian.PutUint32(buff[pos+20:pos+24], uint32(len(submap.peers)))
		pos += 24

		for id, peer := range submap.peers {
			copy(buff[pos:pos+20], id[:])
			copy(buff[pos+20:pos+36], (*(*[peersz]byte)(unsafe.Pointer(peer)))[:])
			pos += 36
		}
	}
	db.mu.Unlock()

	return buff, nil
}

func (db *Memory) decodeBinaryUnsafe(data []byte) (peers, hashes int, err error) {
	defer func() {
		// oob slice panics are fine - we keeping loading peers till we panic
		if x := recover(); x != nil {
			err = x.(error)

			if strings.Contains(err.Error(), "slice bounds out of range") {
				err = nil
			}
		}
	}()

	db.make()
	var pos int

	for {
		var hash storage.Hash
		var count uint32

		copy(hash[:], data[pos:pos+20])
		copy((*(*[4]byte)(unsafe.Pointer(&count)))[:], data[pos+20:pos+24])
		pos += 24

		peermap := db.makePeermap(&hash)

		for ; count > 0; count-- {
			var id storage.PeerID
			var peer storage.Peer

			copy(id[:], data[pos:pos+20])
			copy((*(*[peersz]byte)(unsafe.Pointer(&peer)))[:], data[pos+20:pos+36])
			pos += 36

			peermap.peers[id] = &peer
			peers++
		}

		hashes++
	}
}

package gomap

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/storage"
)

func (db *Memory) encodeBinary() ([]byte, error) {
	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)

	db.mutex.RLock()
	for hash, submap := range db.hashmap {
		db.mutex.RUnlock()

		// write hash and peermap size
		if err := binary.Write(writer, binary.LittleEndian, &hash); err != nil {
			return nil, err
		}
		if err := binary.Write(writer, binary.LittleEndian, uint32(len(submap.Peers))); err != nil {
			return nil, err
		}

		// write peerid and peer
		submap.mutex.RLock()
		for id, peer := range submap.Peers {
			if err := binary.Write(writer, binary.LittleEndian, &id); err != nil {
				return nil, err
			}

			addrSlice := peer.IP.AsSlice()
			if err := binary.Write(writer, binary.LittleEndian, peer.Complete); err != nil {
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, int32(len(addrSlice))); err != nil {
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, addrSlice); err != nil {
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, peer.Port); err != nil {
				return nil, err
			}
			if err := binary.Write(writer, binary.LittleEndian, peer.LastSeen); err != nil {
				return nil, err
			}
		}
		submap.mutex.RUnlock()

		db.mutex.RLock()
	}
	db.mutex.RUnlock()

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (db *Memory) decodeBinary(data []byte) (peers, hashes int, err error) {
	db.make()
	reader := bufio.NewReader(bytes.NewBuffer(data))

	for {
		// decode hash and number of peers
		var hash storage.Hash
		err = binary.Read(reader, binary.LittleEndian, &hash)
		if errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}

		var count uint32
		var complete uint16
		peermap := db.makePeermap(hash)
		if err = binary.Read(reader, binary.LittleEndian, &count); err != nil {
			return
		}

		// decode peerid and peers
		for ; count > 0; count-- {
			var id storage.PeerID
			if err = binary.Read(reader, binary.LittleEndian, &id); err != nil {
				return
			}

			peer := pools.Peers.Get()
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
			peermap.Peers[id] = peer
			peers++

			if peer.Complete {
				complete++
			}
		}

		// set complete and incomplete
		peermap.Complete = complete
		peermap.Incomplete = uint16(len(peermap.Peers)) - complete

		hashes++
	}

	return
}

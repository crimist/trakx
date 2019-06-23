package udp

import (
	"time"
)

type connID struct {
	ID     int64
	cached int64
}

type UDPConnDB map[[4]byte]connID

var connDB UDPConnDB

func (db UDPConnDB) Add(id int64, addr [4]byte) {
	db[addr] = connID{
		ID:     id,
		cached: time.Now().Unix(),
	}
}

func (db UDPConnDB) Check(id int64, addr [4]byte) (ok bool) {
	if id == db[addr].ID {
		ok = true
	}
	return
}

func (db *UDPConnDB) Trim() {
	for key, cID := range connDB {
		if cID.cached+90 < time.Now().Unix() { // older than 90s gets deleted
			delete(connDB, key)
		}
	}
}

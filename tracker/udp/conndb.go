package udp

import (
	"time"

	"github.com/Syc0x00/Trakx/tracker/shared"
	"go.uber.org/zap"
)

var connDB udpConnDB

type connID struct {
	ID     int64
	cached int64
}

type udpConnDB map[[4]byte]connID

func (db udpConnDB) add(id int64, addr [4]byte) {
	if shared.Env == shared.Dev {
		shared.Logger.Info("Add UDPConnDB",
			zap.Int64("ID", id),
		)
	}

	db[addr] = connID{
		ID:     id,
		cached: time.Now().Unix(),
	}
}

func (db udpConnDB) check(id int64, addr [4]byte) (dbID int64, ok bool) {
	if id == db[addr].ID {
		ok = true
	} else {
		dbID = db[addr].ID
	}
	return
}

// Spec says to only cache connIDs for 2min but realistically the chances of it being abused for ddos
// is insanely low so I'll accept them for up to 6 hours
func (db *udpConnDB) trim() {
	trimmed := 0
	now := time.Now().Unix()
	for key, cID := range connDB {
		if now-cID.cached > 21600 { // read note
			delete(connDB, key)
			trimmed++
		}
	}

	shared.Logger.Info("Trim UDPConnDB", zap.Int("trimmed", trimmed), zap.Int("left", len(connDB)))
}

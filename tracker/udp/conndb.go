package udp

import (
	"time"

	"github.com/Syc0x00/Trakx/tracker/shared"
	"go.uber.org/zap"
)

type connID struct {
	ID     int64
	cached int64
}

type UDPConnDB map[[4]byte]connID

var connDB UDPConnDB

func (db UDPConnDB) Add(id int64, addr [4]byte) {
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

func (db UDPConnDB) Check(id int64, addr [4]byte) (ok bool) {
	if id == db[addr].ID {
		ok = true
	}
	return
}

func (db *UDPConnDB) Trim() {
	trimmed := 0
	now := time.Now().Unix()
	for key, cID := range connDB {
		if now-cID.cached > 120 { // older than 2min gets deleted
			delete(connDB, key)
			trimmed++
		}
	}

	shared.Logger.Info("Trim UDPConnDB", zap.Int("trimmed", trimmed), zap.Int("left", len(connDB)))
}

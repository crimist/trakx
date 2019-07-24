package udp

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
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

func WriteConnDB() {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(&connDB); err != nil {
		shared.Logger.Error("conndb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(shared.Config.Database.Conn.Filename, buff.Bytes(), 0644); err != nil {
		shared.Logger.Error("conndb writefile", zap.Error(err))
	}

	shared.Logger.Info("Wrote conndb", zap.Int("entries", len(connDB)))
}

func loadConnDB() {
	file, err := os.Open(shared.Config.Database.Conn.Filename)
	if err != nil {
		shared.Logger.Error("conndb open", zap.Error(err))
		connDB = make(udpConnDB)
		return
	}

	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&connDB); err != nil {
		shared.Logger.Error("conndb decode", zap.Error(err))
		connDB = make(udpConnDB)
		return
	}

	shared.Logger.Info("Loaded conndb", zap.Int("entries", len(connDB)))
}

func (db udpConnDB) add(id int64, addr [4]byte) {
	if shared.Env == shared.Dev {
		shared.Logger.Info("Add conndb",
			zap.Int64("id", id),
			zap.Any("addr", addr),
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

	shared.Logger.Info("Trim conndb", zap.Int("trimmed", trimmed), zap.Int("left", len(connDB)))
}

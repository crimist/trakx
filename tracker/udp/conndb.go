package udp

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

var connDB connectionDatabase

type connID struct {
	ID     int64
	cached int64
}

type connectionDatabase struct {
	mu sync.RWMutex
	db map[[4]byte]connID
}

func (db *connectionDatabase) conns() int {
	return len(db.db)
}

func (db connectionDatabase) add(id int64, addr [4]byte) {
	if !shared.Config.Trakx.Prod {
		shared.Logger.Info("Add conndb",
			zap.Int64("id", id),
			zap.Any("addr", addr),
		)
	}

	db.mu.Lock()
	db.db[addr] = connID{
		ID:     id,
		cached: time.Now().Unix(),
	}
	db.mu.Unlock()
}

func (db connectionDatabase) check(id int64, addr [4]byte) (dbID int64, ok bool) {
	db.mu.RLock()
	dbID = db.db[addr].ID
	ok = id == dbID
	db.mu.RUnlock()
	return
}

// spec says to only cache connIDs for 2min but realistically ips changing for ddos is unlikely so higher can be used
func (db *connectionDatabase) trim() {
	shared.Logger.Info("Trimming connection database")

	start := time.Now()
	epoch := start.Unix()
	trimmed := 0

	db.mu.Lock()
	for key, conn := range db.db {
		if epoch-conn.cached > shared.Config.Database.Conn.Timeout {
			delete(db.db, key)
			trimmed++
		}
	}
	db.mu.Unlock()

	shared.Logger.Info("Trimmed connection database", zap.Int("removed", trimmed), zap.Int("left", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *connectionDatabase) write() {
	shared.Logger.Info("Writing connection database")

	start := time.Now()
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	db.mu.RLock()
	err := encoder.Encode(&db.db)
	db.mu.RUnlock()
	if err != nil {
		shared.Logger.Error("conndb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(shared.Config.Database.Conn.Filename, buff.Bytes(), 0644); err != nil {
		shared.Logger.Error("conndb writefile", zap.Error(err))
	}

	shared.Logger.Info("Wrote connection database", zap.Int("connections", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *connectionDatabase) load() {
	shared.Logger.Info("Loading connection database")
	start := time.Now()

	file, err := os.Open(shared.Config.Database.Conn.Filename)
	if err != nil {
		shared.Logger.Error("conndb open", zap.Error(err))
		db.db = make(map[[4]byte]connID)
		return
	}

	decoder := gob.NewDecoder(file)
	db.mu.Lock()
	err = decoder.Decode(&db.db)
	db.mu.Unlock()
	if err != nil {
		shared.Logger.Error("conndb decode", zap.Error(err))
		db.db = make(map[[4]byte]connID)
		return
	}

	shared.Logger.Info("Loaded connection database", zap.Int("connections", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
}

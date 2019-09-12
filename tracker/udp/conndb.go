package udp

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/syc0x00/trakx/tracker/storage"
	"go.uber.org/zap"
)

const (
	conndbCap = 50000
)

type connID struct {
	ID int64
	ts int64
}

type connectionDatabase struct {
	mu sync.RWMutex
	db map[storage.PeerIP]connID

	timeout  int64
	filename string
	logger   *zap.Logger
}

func newConnectionDatabase(timeout int64, filename string, logger *zap.Logger) *connectionDatabase {
	db := connectionDatabase{
		timeout:  timeout,
		filename: filename,
		logger:   logger,
	}

	if success := db.load(); !success {
		db.make()
	}

	return &db
}

func (db *connectionDatabase) conns() (count int) {
	db.mu.RLock()
	count = len(db.db)
	db.mu.RUnlock()
	return
}

func (db *connectionDatabase) add(id int64, addr storage.PeerIP) {
	db.mu.Lock()
	db.db[addr] = connID{
		ID: id,
		ts: time.Now().Unix(),
	}
	db.mu.Unlock()
}

func (db *connectionDatabase) check(id int64, addr storage.PeerIP) bool {
	db.mu.RLock()
	cid, ok := db.db[addr]
	db.mu.RUnlock()

	if ok && cid.ID == id {
		return true
	}
	return false
}

// spec says to only cache connIDs for 2min but realistically ips changing for ddos is unlikely so higher can be used
func (db *connectionDatabase) trim() {
	db.logger.Info("Trimming connection database")

	start := time.Now()
	epoch := start.Unix()
	trimmed := 0

	db.mu.Lock()
	for key, conn := range db.db {
		if epoch-conn.ts > db.timeout {
			delete(db.db, key)
			trimmed++
		}
	}
	db.mu.Unlock()

	db.logger.Info("Trimmed connection database", zap.Int("removed", trimmed), zap.Int("left", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *connectionDatabase) write() {
	db.logger.Info("Writing connection database")

	start := time.Now()
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	db.trim()

	db.mu.RLock()
	err := encoder.Encode(&db.db)
	db.mu.RUnlock()
	if err != nil {
		db.logger.Error("conndb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(db.filename, buff.Bytes(), 0644); err != nil {
		db.logger.Error("conndb writefile", zap.Error(err))
	}

	db.logger.Info("Wrote connection database", zap.Int("connections", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *connectionDatabase) load() bool {
	db.logger.Info("Loading connection database")
	start := time.Now()

	file, err := os.Open(db.filename)
	if err != nil {
		db.logger.Error("conndb open", zap.Error(err))
		return false
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	db.mu.Lock()
	err = decoder.Decode(&db.db)
	db.mu.Unlock()
	if err != nil {
		db.logger.Error("conndb decode", zap.Error(err))
		return false
	}

	db.logger.Info("Loaded connection database", zap.Int("connections", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
	return true
}

func (db *connectionDatabase) make() {
	db.db = make(map[storage.PeerIP]connID, conndbCap)
}

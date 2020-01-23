package udp

import (
	"io/ioutil"
	"sync"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	conndbAlloc = 50000
	connIDSize  = int(unsafe.Sizeof(connID{}))
)

type connAddr [6]byte
type connID struct {
	ID int64
	ts int64
}

type connectionDatabase struct {
	mu sync.RWMutex
	db map[connAddr]connID

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

	if err := db.load(); err != nil {
		logger.Warn("Failed to load connection database, creating empty db", zap.Error(err))
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

func (db *connectionDatabase) add(id int64, addr connAddr) {
	db.mu.Lock()
	db.db[addr] = connID{
		ID: id,
		ts: time.Now().Unix(),
	}
	db.mu.Unlock()
}

func (db *connectionDatabase) check(id int64, addr connAddr) bool {
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

func (db *connectionDatabase) write() (err error) {
	db.logger.Info("Writing connection database")
	start := time.Now()

	defer func() {
		// recover any oob slice panics
		if tmp := recover(); tmp != nil {
			err = errors.Wrap(tmp.(error), "oob slice panic caught")
		}
	}()

	var pos int
	buff := make([]byte, len(db.db)*22)

	db.mu.Lock()
	for addr, id := range db.db {
		copy(buff[pos:pos+6], addr[:])
		copy(buff[pos+6:pos+22], (*(*[connIDSize]byte)(unsafe.Pointer(&id)))[:])
		pos += 22
	}
	db.mu.Unlock()

	if err := ioutil.WriteFile(db.filename, buff, 0644); err != nil {
		return errors.Wrap(err, "Failed to write connection database to file")
	}

	db.logger.Info("Wrote connection database", zap.Int("connections", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
	return nil
}

func (db *connectionDatabase) load() (err error) {
	db.logger.Info("Loading connection database")
	start := time.Now()

	defer func() {
		// recover any oob slice panics
		if tmp := recover(); tmp != nil {
			err = errors.Wrap(tmp.(error), "oob slice panic caught")
		}
	}()

	data, err := ioutil.ReadFile(db.filename)
	if err != nil {
		return errors.Wrap(err, "failed to read connection database file from disk")
	}

	db.make()

	for pos := 0; pos < len(data); pos += 22 {
		var addr connAddr
		var id connID

		copy(addr[:], data[pos:pos+6])
		copy((*(*[connIDSize]byte)(unsafe.Pointer(&id)))[:], data[pos+6:pos+22])

		db.db[addr] = id
	}

	db.logger.Info("Loaded connection database", zap.Int("connections", db.conns()), zap.Duration("duration", time.Now().Sub(start)))
	return nil
}

func (db *connectionDatabase) make() {
	db.db = make(map[connAddr]connID, conndbAlloc)
}

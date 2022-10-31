package udp

import (
	"io/ioutil"
	"net/netip"
	"sync"
	"time"
	"unsafe"

	"github.com/crimist/trakx/tracker/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	connectionIdSize = int(unsafe.Sizeof(connectionInfo{}))
	addrPortSize     = 26 // netip.Addr + uint16 = 24 + 2
	entrySize        = connectionIdSize + addrPortSize
)

type connectionInfo struct {
	id        int64
	timeStamp int64
}

type connectionDatabase struct {
	mutex         sync.RWMutex
	connectionMap map[netip.AddrPort]connectionInfo
	expiry        int64
}

func newConnectionDatabase(expiry time.Duration) *connectionDatabase {
	connDb := connectionDatabase{
		expiry: int64(expiry.Seconds()),
	}

	connDb.make()

	return &connDb
}

func (db *connectionDatabase) size() (count int) {
	db.mutex.RLock()
	count = len(db.connectionMap)
	db.mutex.RUnlock()
	return
}

func (db *connectionDatabase) add(id int64, addr netip.AddrPort) {
	db.mutex.Lock()
	db.connectionMap[addr] = connectionInfo{
		id:        id,
		timeStamp: time.Now().Unix(),
	}
	db.mutex.Unlock()
}

func (db *connectionDatabase) check(id int64, addr netip.AddrPort) bool {
	db.mutex.RLock()
	cid, ok := db.connectionMap[addr]
	db.mutex.RUnlock()

	if ok && cid.id == id {
		return true
	}
	return false
}

// spec says to only cache connIDs for 2min but realistically ips changing for ddos is unlikely so higher can be used
func (db *connectionDatabase) trim() {
	config.Logger.Info("Trimming connection database")

	start := time.Now()
	epoch := start.Unix()
	trimmed := 0

	db.mutex.Lock()
	for key, conn := range db.connectionMap {
		if epoch-conn.timeStamp > db.expiry {
			delete(db.connectionMap, key)
			trimmed++
		}
	}
	db.mutex.Unlock()

	config.Logger.Info("Trimmed connection database", zap.Int("removed", trimmed), zap.Int("left", db.size()), zap.Duration("duration", time.Since(start)))
}

func (connDb *connectionDatabase) writeToFile(path string) error {
	config.Logger.Info("Writing connection database")
	start := time.Now()

	encoded, err := connDb.marshallBinary()
	if err != nil {
		return errors.Wrap(err, "failed to marshall connection database")
	}

	if err := ioutil.WriteFile(path, encoded, 0644); err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	config.Logger.Info("Wrote connection database", zap.Int("connections", connDb.size()), zap.Duration("duration", time.Since(start)))
	return nil

}

func (db *connectionDatabase) loadFromFile(path string) error {
	config.Logger.Info("Loading connection database")
	start := time.Now()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to read connection database file from disk")
	}

	if err := db.unmarshallBinary(data); err != nil {
		return errors.Wrap(err, "failed to unmarshall binary data")

	}

	config.Logger.Info("Loaded connection database", zap.Int("connections", db.size()), zap.Duration("duration", time.Since(start)))
	return nil
}

func (db *connectionDatabase) marshallBinary() (buff []byte, err error) {
	defer func() {
		// recover any oob slice panics
		if tmp := recover(); tmp != nil {
			err = errors.Wrap(tmp.(error), "oob slice panic caught")
		}
	}()

	var pos int
	buff = make([]byte, len(db.connectionMap)*entrySize)

	db.mutex.Lock()
	for addr, id := range db.connectionMap {
		copy(buff[pos:pos+addrPortSize], (*(*[addrPortSize]byte)(unsafe.Pointer(&addr)))[:])
		copy(buff[pos+addrPortSize:pos+entrySize], (*(*[connectionIdSize]byte)(unsafe.Pointer(&id)))[:])
		pos += entrySize
	}
	db.mutex.Unlock()

	return buff, nil
}

func (db *connectionDatabase) unmarshallBinary(data []byte) (err error) {
	defer func() {
		// recover any oob slice panics
		if tmp := recover(); tmp != nil {
			err = errors.Wrap(tmp.(error), "oob slice panic caught")
		}
	}()

	db.make()

	for pos := 0; pos < len(data); pos += entrySize {
		var addr netip.AddrPort
		var id connectionInfo

		copy((*(*[addrPortSize]byte)(unsafe.Pointer(&addr)))[:], data[pos:pos+addrPortSize])
		copy((*(*[connectionIdSize]byte)(unsafe.Pointer(&id)))[:], data[pos+addrPortSize:pos+entrySize])

		db.connectionMap[addr] = id
	}

	return nil
}

func (db *connectionDatabase) make() {
	db.connectionMap = make(map[netip.AddrPort]connectionInfo, config.Config.UDP.ConnDB.Size)
}

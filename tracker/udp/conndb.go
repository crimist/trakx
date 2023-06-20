package udp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/crimist/trakx/tracker/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type connectionInfo struct {
	ID        int64
	TimeStamp int64
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
		ID:        id,
		TimeStamp: time.Now().Unix(),
	}
	db.mutex.Unlock()
}

func (db *connectionDatabase) check(id int64, addr netip.AddrPort) bool {
	db.mutex.RLock()
	cid, ok := db.connectionMap[addr]
	db.mutex.RUnlock()

	if ok && cid.ID == id {
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
		if epoch-conn.TimeStamp > db.expiry {
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

	encoded, err := connDb.gobEncode()
	if err != nil {
		return errors.Wrap(err, "failed to marshall connection database")
	}

	if err := os.WriteFile(path, encoded, 0644); err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	config.Logger.Info("Wrote connection database", zap.Int("connections", connDb.size()), zap.Duration("duration", time.Since(start)))
	return nil

}

func (db *connectionDatabase) loadFromFile(path string) error {
	config.Logger.Info("Loading connection database")
	start := time.Now()

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to read connection database file from disk")
	}

	if err := db.gobDecode(data); err != nil {
		return errors.Wrap(err, "failed to unmarshall binary data")

	}

	config.Logger.Info("Loaded connection database", zap.Int("connections", db.size()), zap.Duration("duration", time.Since(start)))
	return nil
}

func (db *connectionDatabase) gobEncode() ([]byte, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	db.mutex.Lock()
	gob.NewEncoder(writer).Encode(db.connectionMap)
	db.mutex.Unlock()

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (db *connectionDatabase) gobDecode(data []byte) error {
	db.make()
	reader := bufio.NewReader(bytes.NewBuffer(data))

	return gob.NewDecoder(reader).Decode(&db.connectionMap)
}

func (db *connectionDatabase) marshallBinary() ([]byte, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	db.mutex.Lock()
	if err := binary.Write(writer, binary.LittleEndian, uint32(len(db.connectionMap))); err != nil {
		return nil, err
	}
	for addrPort, connInfo := range db.connectionMap {
		addrSlice := addrPort.Addr().AsSlice()

		if err := binary.Write(writer, binary.LittleEndian, uint8(len(addrSlice))); err != nil {
			return nil, err
		}
		if err := binary.Write(writer, binary.LittleEndian, addrSlice); err != nil {
			return nil, err
		}
		if err := binary.Write(writer, binary.LittleEndian, addrPort.Port()); err != nil {
			return nil, err
		}
		if err := binary.Write(writer, binary.LittleEndian, connInfo); err != nil {
			return nil, err
		}
	}
	db.mutex.Unlock()

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (db *connectionDatabase) unmarshallBinary(data []byte) error {
	db.make()
	reader := bufio.NewReader(bytes.NewBuffer(data))

	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return err
	}

	for ; length > 0; length-- {
		var addrSliceLen uint8
		var port uint16
		var connInfo connectionInfo

		if err := binary.Read(reader, binary.LittleEndian, &addrSliceLen); err != nil {
			return err
		}
		addrSlice := make([]byte, addrSliceLen)
		if err := binary.Read(reader, binary.LittleEndian, &addrSlice); err != nil {
			return err
		}
		if err := binary.Read(reader, binary.LittleEndian, &port); err != nil {
			return err
		}
		if err := binary.Read(reader, binary.LittleEndian, &connInfo); err != nil {
			return err
		}

		addr, ok := netip.AddrFromSlice(addrSlice)
		if !ok {
			return errors.New("failed to parse addr from slice")
		}
		db.connectionMap[netip.AddrPortFrom(addr, port)] = connInfo
	}

	return nil
}

func (db *connectionDatabase) make() {
	db.connectionMap = make(map[netip.AddrPort]connectionInfo, config.Config.UDP.ConnDB.Size)
}

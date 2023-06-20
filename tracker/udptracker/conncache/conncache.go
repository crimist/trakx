package conncache

import (
	"net/netip"
	"sync"
	"time"

	"github.com/crimist/trakx/utils"
	"go.uber.org/zap"
)

type entry struct {
	ID        int64
	TimeStamp int64
}

type entryMap map[netip.AddrPort]entry

type ConnectionCache struct {
	mutex          sync.RWMutex
	entries        entryMap
	expirationTime int64
}

func NewConnectionCache(initalEntrySlots int, expirationTime time.Duration, cleanupFrequency time.Duration, persistancePath string) *ConnectionCache {
	var err error
	connCache := ConnectionCache{
		expirationTime: int64(expirationTime.Seconds()),
	}

	if persistancePath != "" {
		connCache.entries, err = loadEntriesFromFile(persistancePath)
		if err != nil {
			zap.L().Warn("failed to load entries from file, initializing empty connection cache", zap.Error(err))
			connCache.entries = make(entryMap, initalEntrySlots)
		}
	} else {
		connCache.entries = make(entryMap, initalEntrySlots)
	}

	go utils.RunOn(cleanupFrequency, connCache.evictExpiredEntries)

	return &connCache
}

func (connCache *ConnectionCache) EntryCount() int {
	connCache.mutex.RLock()
	count := len(connCache.entries)
	connCache.mutex.RUnlock()
	return count
}

func (connCache *ConnectionCache) Set(id int64, addr netip.AddrPort) {
	connCache.mutex.Lock()
	connCache.entries[addr] = entry{
		ID:        id,
		TimeStamp: time.Now().Unix(),
	}
	connCache.mutex.Unlock()
}

func (connCache *ConnectionCache) Validate(id int64, addr netip.AddrPort) bool {
	connCache.mutex.RLock()
	cid, ok := connCache.entries[addr]
	connCache.mutex.RUnlock()

	return ok && cid.ID == id
}

// standard says cache connection entryes for 2 minuites. Realistically this is unlikely to change so we can cache longer.
func (db *ConnectionCache) evictExpiredEntries() {
	zap.L().Info("evicting expired entries from connection cache")

	start := time.Now()
	now := start.Unix()
	evicted := 0

	db.mutex.Lock()
	for key, conn := range db.entries {
		if now-conn.TimeStamp > db.expirationTime {
			delete(db.entries, key)
			evicted++
		}
	}
	db.mutex.Unlock()

	zap.L().Info("evicted expired entries from connection cache", zap.Int("evictedcount", evicted), zap.Int("entrycount", db.EntryCount()), zap.Duration("elasped", time.Since(start)))
}

package connections

import (
	"crypto/rand"
	"encoding/binary"
	"net/netip"
	"sync"
	"time"

	"github.com/crimist/trakx/internal/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type associationEntry struct {
	ID        uint64
	TimeStamp int64
}

type associations map[netip.AddrPort]associationEntry

type Connections struct {
	mutex        sync.RWMutex
	associations associations
	maxAge       int64
}

func NewConnections(initialSize int, maxAge time.Duration, gcFrequency time.Duration) *Connections {
	connections := Connections{
		maxAge:       int64(maxAge.Seconds()),
		associations: make(associations, initialSize),
	}

	if gcFrequency > 0 {
		go utils.RunOn(gcFrequency, connections.garbageCollector)
	}

	return &connections
}

func (connCache *Connections) Entries() int {
	connCache.mutex.RLock()
	count := len(connCache.associations)
	connCache.mutex.RUnlock()
	return count
}

func (connCache *Connections) Create(addr netip.AddrPort) (uint64, error) {
	var connectionID uint64
	err := binary.Read(rand.Reader, binary.BigEndian, &connectionID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate connection ID")
	}

	epoch := time.Now().Unix()

	connCache.mutex.Lock()
	connCache.associations[addr] = associationEntry{
		ID:        connectionID,
		TimeStamp: epoch,
	}
	connCache.mutex.Unlock()

	return connectionID, nil
}

func (connCache *Connections) Validate(addr netip.AddrPort, id uint64) bool {
	connCache.mutex.RLock()
	entry, ok := connCache.associations[addr]
	connCache.mutex.RUnlock()

	// it's possible for an expired entry to be validated here if it has yet to be collected
	// but this should be acceptable so long as the gc frequency is reasonable
	return ok && entry.ID == id
}

func (connections *Connections) garbageCollector() {
	zap.L().Debug("beginning connections garbage collector")

	start := time.Now()
	epoch := start.Unix()
	evicted := 0

	connections.mutex.Lock()
	for key, entry := range connections.associations {
		if epoch-entry.TimeStamp > connections.maxAge {
			delete(connections.associations, key)
			evicted++
		}
	}
	connections.mutex.Unlock()

	zap.L().Info("connections garbage collection complete", zap.Int("evicted", evicted), zap.Int("post-entries", connections.Entries()), zap.Duration("elapsed", time.Since(start)))
}

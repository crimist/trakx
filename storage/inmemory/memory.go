/*
	Map implements a trakx database through go maps in local memory. It is heavily optimized for performance but cannot be shared accross multiple trackers as it resides in local memory.
*/

package inmemory

import (
	"sync"
	"time"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type PeerMap struct {
	mutex      sync.RWMutex // can't be embedded (https://github.com/golang/go/issues/5819#issuecomment-250596051)
	Complete   uint16
	Incomplete uint16
	Peers      map[storage.PeerID]*storage.Peer
}

type InMemory struct {
	mutex  sync.RWMutex
	hashes map[storage.Hash]*PeerMap
}

func NewInMemory(config storage.Config) (storage.Database, error) {
	InMemory := InMemory{}
	return InMemory, nil
}

func (db *InMemory) Init(persistance storage.Persistance, config storage.Config) error {
	*db = InMemory{}

	if err := db.backup.Init(db); err != nil {
		return errors.Wrap(err, "failed to initialize backup")
	}
	if err := db.backup.Load(); err != nil {
		return errors.Wrap(err, "failed to load backup")
	}

	if config.Config.DB.Backup.Frequency > 0 {
		go utils.RunOn(config.Config.DB.Backup.Frequency, func() {
			if err := db.backup.Save(); err != nil {
				zap.L().Info("Failed to backup the database", zap.Error(err))
			}
		})
	}
	if config.Config.DB.Trim > 0 {
		go utils.RunOn(config.Config.DB.Trim, db.Trim)
	}

	return nil
}

func (db *InMemory) make() {
	db.hashes = make(map[storage.Hash]*PeerMap, prealloactedHashes)
}

func (db *InMemory) makePeermap(h storage.Hash) (peermap *PeerMap) {
	// build struct and assign
	peermap = new(PeerMap)
	peermap.Peers = make(map[storage.PeerID]*storage.Peer, peerMapPrealloc)
	db.hashes[h] = peermap
	return
}

func (db *InMemory) Backup() storage.Backup {
	return db.backup
}

func (db *InMemory) Trim() {
	start := time.Now()
	zap.L().Info("Trimming database")
	peers, hashes := db.trim()
	zap.L().Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Since(start)))
}

func (db *InMemory) trim() (peers, hashes int) {
	now := time.Now().Unix()
	peerTimeout := int64(config.Config.DB.Expiry.Seconds())

	db.mutex.RLock()
	for hash, peermap := range db.hashes {
		db.mutex.RUnlock()

		peermap.mutex.Lock()
		for id, peer := range peermap.Peers {
			if now-peer.LastSeen > peerTimeout {
				db.delete(peer, peermap, id)
				peers++
			}
		}
		peersize := len(peermap.Peers)
		peermap.mutex.Unlock()

		if peersize == 0 {
			db.mutex.Lock()
			delete(db.hashes, hash)
			db.mutex.Unlock()
			hashes++
		}
		db.mutex.RLock()
	}
	db.mutex.RUnlock()

	return
}

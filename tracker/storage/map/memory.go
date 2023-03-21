/*
	Map implements a trakx database through go maps in local memory. It is heavily optimized for performance but cannot be shared accross multiple trackers as it resides in local memory.
*/

package gomap

import (
	"sync"
	"time"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	hashMapPrealloc = 250_000
	peerMapPrealloc = 1
)

type PeerMap struct {
	mutex      sync.RWMutex // can't be embedded (https://github.com/golang/go/issues/5819#issuecomment-250596051)
	Complete   uint16
	Incomplete uint16
	Peers      map[storage.PeerID]*storage.Peer
}

type Memory struct {
	mutex   sync.RWMutex
	hashmap map[storage.Hash]*PeerMap

	backup storage.Backup
}

func (db *Memory) Init(backup storage.Backup) error {
	*db = Memory{
		backup: backup,
	}

	if err := db.backup.Init(db); err != nil {
		return errors.Wrap(err, "failed to initialize backup")
	}
	if err := db.backup.Load(); err != nil {
		return errors.Wrap(err, "failed to load backup")
	}

	if config.Config.DB.Backup.Frequency > 0 {
		go utils.RunOn(config.Config.DB.Backup.Frequency, func() {
			if err := db.backup.Save(); err != nil {
				config.Logger.Info("Failed to backup the database", zap.Error(err))
			}
		})
	}
	if config.Config.DB.Trim > 0 {
		go utils.RunOn(config.Config.DB.Trim, db.Trim)
	}

	return nil
}

func (db *Memory) make() {
	db.hashmap = make(map[storage.Hash]*PeerMap, hashMapPrealloc)
}

func (db *Memory) makePeermap(h storage.Hash) (peermap *PeerMap) {
	// build struct and assign
	peermap = new(PeerMap)
	peermap.Peers = make(map[storage.PeerID]*storage.Peer, peerMapPrealloc)
	db.hashmap[h] = peermap
	return
}

func (db *Memory) Backup() storage.Backup {
	return db.backup
}

func (db *Memory) Check() bool {
	return db.hashmap != nil
}

func (db *Memory) Trim() {
	start := time.Now()
	config.Logger.Info("Trimming database")
	peers, hashes := db.trim()
	config.Logger.Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Since(start)))
}

func (db *Memory) trim() (peers, hashes int) {
	now := time.Now().Unix()
	peerTimeout := int64(config.Config.DB.Expiry.Seconds())

	db.mutex.RLock()
	for hash, peermap := range db.hashmap {
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
			delete(db.hashmap, hash)
			db.mutex.Unlock()
			hashes++
		}
		db.mutex.RLock()
	}
	db.mutex.RUnlock()

	return
}

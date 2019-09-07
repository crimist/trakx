package inmemory

import (
	"sync"
	"time"

	"github.com/syc0x00/trakx/tracker/database"
	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

// nmber of hashes to make()
const initCap = 1000000

type subPeerMap struct {
	sync.RWMutex
	peers map[shared.PeerID]*shared.Peer
}

type Memory struct {
	mu      sync.RWMutex
	hashmap map[shared.Hash]*subPeerMap

	backup database.Backup
	conf   *shared.Config
	logger *zap.Logger
}

func (db *Memory) Init(conf *shared.Config, logger *zap.Logger, backup database.Backup) {
	*db = Memory{
		conf:   conf,
		logger: logger,
		backup: backup,
	}

	if err := db.backup.Init(db); err != nil {
		panic(err)
	}
	if err := db.backup.Load(); err != nil {
		panic(err)
	}

	if conf.Database.Peer.Write > 0 {
		go shared.RunOn(time.Duration(conf.Database.Peer.Write)*time.Second, func() {
			db.backup.SaveTmp()
		})
	}
	if conf.Database.Peer.Trim > 0 {
		go shared.RunOn(time.Duration(conf.Database.Peer.Trim)*time.Second, db.Trim)
	}
}

func (db *Memory) make() { db.hashmap = make(map[shared.Hash]*subPeerMap, initCap) }

func (db *Memory) Backup() database.Backup { return db.backup }

func (db *Memory) Check() bool { return db.hashmap != nil }

func (db *Memory) Trim() {
	start := time.Now()
	db.logger.Info("Trimming database")
	peers, hashes := db.trim()
	if peers < 1 && hashes < 1 {
		db.logger.Info("Can't trim database: database empty")
	} else {
		db.logger.Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Now().Sub(start)))
	}
}

func (db *Memory) trim() (peers, hashes int) {
	now := time.Now().Unix()

	db.mu.Lock()
	for hash, peermap := range db.hashmap {
		db.mu.Unlock()

		peermap.Lock()
		for id, peer := range peermap.peers {
			if now-peer.LastSeen > db.conf.Database.Peer.Timeout {
				db.delete(peer, &hash, &id)
				peers++
			}
		}
		peermap.Unlock()

		db.mu.Lock()
		if len(peermap.peers) == 0 {
			delete(db.hashmap, hash)
			hashes++
		}
	}
	db.mu.Unlock()

	return
}

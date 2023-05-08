/*
	Map implements a trakx database through go maps in local memory. It is heavily optimized for performance but cannot be shared accross multiple trackers as it resides in local memory.
*/

package inmemory

import (
	"sync"
	"time"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/utils"
	"go.uber.org/zap"
)

const torrentPeerPrealloc = 1

type Torrent struct {
	mutex   sync.RWMutex // can't be embedded (https://github.com/golang/go/issues/5819#issuecomment-250596051)
	Seeds   uint16
	Leeches uint16
	Peers   map[storage.PeerID]*storage.Peer
}

type InMemory struct {
	mutex    sync.RWMutex
	torrents map[storage.Hash]*Torrent
	stats    *stats.Statistics
	peerPool *pools.Pool[*storage.Peer]
}

func NewInMemory(config Config) (*InMemory, error) {
	db := &InMemory{
		stats: config.Stats,
		peerPool: pools.NewPool[*storage.Peer](func() any {
			return new(storage.Peer)
		}, nil),
	}

	if config.Persistance != nil {
		if err := config.Persistance.read(db, config.PersistanceAddress); err != nil {
			zap.L().Warn("Failed to load database from persistance", zap.Any("persistance", config.Persistance), zap.String("address", config.PersistanceAddress), zap.Error(err))
			db.torrents = make(map[storage.Hash]*Torrent, config.InitalSize)
		} else {
			zap.L().Info("Loaded database from persistance", zap.Any("persistance", config.Persistance), zap.String("address", config.PersistanceAddress), zap.Int("torrents", db.Torrents()))
		}
	} else {
		db.torrents = make(map[storage.Hash]*Torrent, config.InitalSize)
	}

	// TODO: refactor the stats package
	db.syncExpvars()

	if config.EvictionFrequency > 0 {
		go utils.RunOn(config.EvictionFrequency, func() {
			db.evictExpired(int64(config.ExpirationTime.Seconds()))
		})
	}

	return db, nil
}

// Torrents returns the number of torrents registered in the database
func (db *InMemory) Torrents() int {
	db.mutex.RLock()
	torrents := len(db.torrents)
	db.mutex.RUnlock()
	return torrents
}

func (db *InMemory) createTorrent(h storage.Hash) *Torrent {
	torrent := new(Torrent)
	torrent.Peers = make(map[storage.PeerID]*storage.Peer, torrentPeerPrealloc)

	db.mutex.Lock()
	db.torrents[h] = torrent
	db.mutex.Unlock()

	return torrent
}

func (db *InMemory) evictExpired(expirationTime int64) {
	now := time.Now()
	zap.L().Info("trimming inmemory database")

	trimmedPeers, trimmedTorrents := 0, 0
	nowUnix := now.Unix()

	db.mutex.RLock()
	for hash, torrent := range db.torrents {
		db.mutex.RUnlock()

		torrent.mutex.Lock()
		for id, peer := range torrent.Peers {
			if nowUnix-peer.LastSeen > expirationTime {
				delete(torrent.Peers, id)

				if peer.Complete {
					torrent.Seeds--
				} else {
					torrent.Leeches--
				}

				if dbStats && db.stats != nil {
					if peer.Complete {
						db.stats.Seeds.Add(-1)
					} else {
						db.stats.Leeches.Add(-1)
					}

					db.stats.IPStats.Lock()
					db.stats.IPStats.Remove(peer.IP)
					db.stats.IPStats.Unlock()
				}

				db.peerPool.Put(peer)
				trimmedPeers++
			}
		}
		numPeers := len(torrent.Peers)
		torrent.mutex.Unlock()

		if numPeers == 0 {
			db.mutex.Lock()
			delete(db.torrents, hash)
			db.mutex.Unlock()
			trimmedTorrents++
		}

		db.mutex.RLock()
	}
	db.mutex.RUnlock()

	zap.L().Info("trimmed inmemory database", zap.Int("peers", trimmedPeers), zap.Int("torrents", trimmedTorrents), zap.Duration("elapsed", time.Since(now)))
}

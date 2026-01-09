/*
	Map implements a trakx database through go maps in local memory. It is heavily optimized for performance but cannot be shared accross multiple trackers as it resides in local memory.
*/

package database

import (
	"io"
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

type Database struct {
	mutex     sync.RWMutex
	torrents  map[storage.Hash]*Torrent
	collector stats.Collector
	peerPool  *pools.Pool[*storage.Peer]
}

func NewDatabase(config Config) (*Database, error) {
	db := &Database{
		collector: config.Collector,
		peerPool: pools.NewPool[*storage.Peer](func() any {
			return new(storage.Peer)
		}, nil),
	}

	if config.ImportReader != nil {
		if err := db.Restore(config.ImportReader); err != nil {
			zap.L().Warn("Failed to restore database from import stream", zap.Error(err))
		} else {
			zap.L().Info("Loaded database from import stream", zap.Int("torrents", db.Torrents()))
		}
		if closer, ok := config.ImportReader.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				zap.L().Warn("Failed to close import reader", zap.Error(err))
			}
		}
	}

	if db.torrents == nil {
		db.torrents = make(map[storage.Hash]*Torrent, config.InitalSize)
	}

	var seeds, leeches int64

	for _, peermap := range db.torrents {
		for _, peer := range peermap.Peers {
			db.collector.IPs().Inc(peer.IP)

			if peer.Complete {
				seeds++
			} else {
				leeches++
			}
		}
	}

	db.collector.AddSeeds(seeds)
	db.collector.AddLeeches(leeches)

	if config.EvictionFrequency > 0 {
		go utils.RunOn(config.EvictionFrequency, func() {
			db.evictExpired(int64(config.ExpirationTime.Seconds()))
		})
	}

	return db, nil
}

// Torrents returns the number of torrents registered in the database
func (db *Database) Torrents() int {
	db.mutex.RLock()
	torrents := len(db.torrents)
	db.mutex.RUnlock()
	return torrents
}

func (db *Database) createTorrent(h storage.Hash) *Torrent {
	torrent := new(Torrent)
	torrent.Peers = make(map[storage.PeerID]*storage.Peer, torrentPeerPrealloc)

	db.mutex.Lock()
	db.torrents[h] = torrent
	db.mutex.Unlock()

	return torrent
}

func (db *Database) evictExpired(expirationTime int64) {
	now := time.Now()
	zap.L().Info("trimming in-memory database")

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
					db.collector.AddSeeds(-1)
				} else {
					torrent.Leeches--
					db.collector.AddLeeches(-1)
				}
				db.collector.IPs().Remove(peer.IP)

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

	zap.L().Info("trimmed in-memory database", zap.Int("peers", trimmedPeers), zap.Int("torrents", trimmedTorrents), zap.Duration("elapsed", time.Since(now)))
}

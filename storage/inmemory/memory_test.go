package inmemory

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/crimist/trakx/storage"
)

func TestNewInMemory(t *testing.T) {
	_, err := NewInMemory(Config{
		InitalSize:         1,
		Persistance:        nil,
		PersistanceAddress: "",
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Minute,
		Stats:              nil,
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}
}

func TestTorrents(t *testing.T) {
	db, err := NewInMemory(Config{
		InitalSize:         1,
		Persistance:        nil,
		PersistanceAddress: "",
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Microsecond,
		Stats:              nil,
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}
	var hash storage.Hash
	var peerid storage.PeerID
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	rnd.Read(hash[:])
	rnd.Read(peerid[:])
	db.PeerAdd(hash, peerid, testPeerIP, 1234, true)
	rnd.Read(peerid[:])
	db.PeerAdd(hash, peerid, testPeerIP, 1234, true)
	rnd.Read(hash[:])
	rnd.Read(peerid[:])
	db.PeerAdd(hash, peerid, testPeerIP, 1234, true)
	torrents := db.Torrents()

	if torrents != 2 {
		t.Errorf("torrents count = %v, want 2", torrents)
	}
}

func TestEviction(t *testing.T) {
	db, err := NewInMemory(Config{
		InitalSize:         1,
		Persistance:        nil,
		PersistanceAddress: "",
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Microsecond,
		Stats:              nil,
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}

	db.PeerAdd(testTorrentHash1, testPeerID1, testPeerIP, 1234, true)
	db.PeerAdd(testTorrentHash2, testPeerID1, testPeerIP, 1234, true)
	time.Sleep(1 * time.Second)
	db.PeerAdd(testTorrentHash2, testPeerID2, testPeerIP, 1234, true)
	db.evictExpired(0)

	_, ok := db.torrents[testTorrentHash1]
	if ok {
		t.Error("empty torrent not evicted from database")
	}
	_, ok = db.torrents[testTorrentHash2].Peers[testPeerID1]
	if ok {
		t.Error("expired peer not evicted from database")
	}
	_, ok = db.torrents[testTorrentHash2].Peers[testPeerID2]
	if !ok {
		t.Error("unexpired peer evicted from database")
	}
}

func BenchmarkEvictionSingle(b *testing.B) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for peers := 1000; peers < 1e7; peers *= 10 {
		b.Run(fmt.Sprintf("%d", peers), func(b *testing.B) {
			b.StopTimer()
			b.ResetTimer()

			db, err := NewInMemory(Config{
				InitalSize:         1,
				Persistance:        nil,
				PersistanceAddress: "",
				EvictionFrequency:  1 * time.Hour,
				ExpirationTime:     1 * time.Hour,
				Stats:              nil,
			})
			if err != nil {
				b.Fatal("Failed to create database")
			}
			var peerid storage.PeerID

			for n := 0; n < b.N; n++ {
				for i := 0; i < peers; i++ {
					rnd.Read(peerid[:])
					db.PeerAdd(testTorrentHash1, peerid, testPeerIP, 1234, true)
				}

				b.StartTimer()
				db.evictExpired(-1)
				b.StopTimer()
			}
		})
	}
}

func BenchmarkEvictionMulti(b *testing.B) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for peers := 1000; peers < 1e7; peers *= 10 {
		b.Run(fmt.Sprintf("%d", peers), func(b *testing.B) {
			b.StopTimer()
			b.ResetTimer()

			db, err := NewInMemory(Config{
				InitalSize:         1,
				Persistance:        nil,
				PersistanceAddress: "",
				EvictionFrequency:  1 * time.Hour,
				ExpirationTime:     1 * time.Hour,
				Stats:              nil,
			})
			if err != nil {
				b.Fatal("Failed to create database")
			}
			var hash storage.Hash
			var peerid storage.PeerID

			for n := 0; n < b.N; n++ {
				for i := 0; i < peers; i++ {
					rnd.Read(peerid[:])
					rnd.Read(hash[:])
					db.PeerAdd(hash, peerid, testPeerIP, 1234, true)
				}

				b.StartTimer()
				db.evictExpired(-1)
				b.StopTimer()
			}
		})
	}
}

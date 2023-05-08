package inmemory

import (
	"fmt"
	"math/rand"
	"net/netip"
	"testing"
	"time"

	"github.com/crimist/trakx/storage"
)

var (
	testTorrentHash1 = storage.Hash([20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	testTorrentHash2 = storage.Hash([20]byte{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	testPeerIP       = netip.AddrFrom4([4]byte{1, 2, 3, 4})
	testPeerID1      = storage.PeerID([20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	testPeerID2      = storage.PeerID([20]byte{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
)

func TestPeerAdd(t *testing.T) {
	db, err := NewInMemory(Config{
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
	testPeer := storage.Peer{
		Complete: true,
		IP:       testPeerIP,
		Port:     1234,
	}

	nowUnix := time.Now().Unix()
	db.PeerAdd(testTorrentHash1, testPeerID1, testPeer.IP, testPeer.Port, testPeer.Complete)

	dbPeer, ok := db.torrents[testTorrentHash1].Peers[testPeerID1]
	if !ok {
		t.Error("peer not added to database")
	}
	if dbPeer.Complete != testPeer.Complete {
		t.Errorf("peer complete = %v, want %v", dbPeer.Complete, testPeer.Complete)
	}
	if dbPeer.IP != testPeer.IP {
		t.Errorf("peer ip = %v, want %v", dbPeer.IP, testPeer.IP)
	}
	if dbPeer.Port != testPeer.Port {
		t.Errorf("peer port = %v, want %v", dbPeer.Port, testPeer.Port)
	}
	if dbPeer.LastSeen != nowUnix {
		t.Errorf("peer lastseen = %v, want %v", dbPeer.LastSeen, nowUnix)
	}
}

func TestPeerRemove(t *testing.T) {
	db, err := NewInMemory(Config{
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
	testPeer := storage.Peer{
		Complete: true,
		IP:       testPeerIP,
		Port:     1234,
	}

	db.PeerAdd(testTorrentHash1, testPeerID1, testPeer.IP, testPeer.Port, testPeer.Complete)
	db.PeerRemove(testTorrentHash1, testPeerID1)

	_, ok := db.torrents[testTorrentHash1].Peers[testPeerID1]
	if ok {
		t.Error("peer not removed from database")
	}
}

func BenchmarkPeerAddSingle(b *testing.B) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	db, err := NewInMemory(Config{
		InitalSize:         1,
		Persistance:        nil,
		PersistanceAddress: "",
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Minute,
		Stats:              nil,
	})
	if err != nil {
		b.Fatal("Failed to create database")
	}
	benchPeer := storage.Peer{
		Complete: true,
		IP:       testPeerIP,
		Port:     1234,
	}
	var peerid storage.PeerID

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rnd.Read(peerid[:])
		db.PeerAdd(testTorrentHash1, peerid, benchPeer.IP, benchPeer.Port, benchPeer.Complete)
	}
}

func BenchmarkPeerAddSingleParallell(b *testing.B) {
	for routines := 1; routines < 1000; routines *= 10 {
		b.Run(fmt.Sprintf("%d", routines), func(b *testing.B) {
			db, err := NewInMemory(Config{
				InitalSize:         1,
				Persistance:        nil,
				PersistanceAddress: "",
				EvictionFrequency:  1 * time.Minute,
				ExpirationTime:     1 * time.Minute,
				Stats:              nil,
			})
			if err != nil {
				b.Fatal("Failed to create database")
			}
			benchPeer := storage.Peer{
				Complete: true,
				IP:       testPeerIP,
				Port:     1234,
			}

			b.SetParallelism(routines)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
				var peerid storage.PeerID

				for pb.Next() {
					rnd.Read(peerid[:])
					db.PeerAdd(testTorrentHash1, peerid, benchPeer.IP, benchPeer.Port, benchPeer.Complete)
				}
			})
		})
	}
}

func BenchmarkPeerAddMulti(b *testing.B) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	db, err := NewInMemory(Config{
		InitalSize:         1,
		Persistance:        nil,
		PersistanceAddress: "",
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Minute,
		Stats:              nil,
	})
	if err != nil {
		b.Fatal("Failed to create database")
	}
	benchPeer := storage.Peer{
		Complete: true,
		IP:       testPeerIP,
		Port:     1234,
	}
	var peerid storage.PeerID
	var hash storage.Hash

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rnd.Read(peerid[:])
		rnd.Read(hash[:])
		db.PeerAdd(hash, peerid, benchPeer.IP, benchPeer.Port, benchPeer.Complete)
	}
}

func BenchmarkPeerAddMultiParallell(b *testing.B) {
	for routines := 1; routines < 1000; routines *= 10 {
		b.Run(fmt.Sprintf("%d", routines), func(b *testing.B) {
			db, err := NewInMemory(Config{
				InitalSize:         1,
				Persistance:        nil,
				PersistanceAddress: "",
				EvictionFrequency:  1 * time.Minute,
				ExpirationTime:     1 * time.Minute,
				Stats:              nil,
			})
			if err != nil {
				b.Fatal("Failed to create database")
			}
			benchPeer := storage.Peer{
				Complete: true,
				IP:       testPeerIP,
				Port:     1234,
			}

			b.SetParallelism(routines)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
				var peerid storage.PeerID
				var hash storage.Hash

				for pb.Next() {
					rnd.Read(peerid[:])
					rnd.Read(hash[:])
					db.PeerAdd(hash, peerid, benchPeer.IP, benchPeer.Port, benchPeer.Complete)
				}
			})
		})
	}
}

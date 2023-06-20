package inmemory

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/crimist/trakx/storage"
)

func TestBinaryCoder(t *testing.T) {
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

	data, err := encodeBinary(db)
	if err != nil {
		t.Fatal("encodeBinary threw error: ", err)
	}
	oldtorrents := db.torrents
	db, err = NewInMemory(Config{
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
	if _, _, err = decodeBinary(db, data); err != nil {
		t.Fatal("decodeBinary threw error: ", err)
	}

	if _, ok := db.torrents[testTorrentHash1]; !ok {
		t.Fatal("torrent missing peer")
	}
	if db.torrents[testTorrentHash1].Seeds != oldtorrents[testTorrentHash1].Seeds {
		t.Fatalf("seeds = %v, want %v", db.torrents[testTorrentHash1].Seeds, oldtorrents[testTorrentHash1].Seeds)
	}
	if db.torrents[testTorrentHash1].Leeches != oldtorrents[testTorrentHash1].Leeches {
		t.Fatalf("leeches = %v, want %v", db.torrents[testTorrentHash1].Leeches, oldtorrents[testTorrentHash1].Leeches)
	}
	if !reflect.DeepEqual(db.torrents[testTorrentHash1].Peers, oldtorrents[testTorrentHash1].Peers) {
		t.Fatalf("peers = %v, want %v", db.torrents[testTorrentHash1].Peers, oldtorrents[testTorrentHash1].Peers)
	}
}

func BenchmarkEncodeBinary(b *testing.B) {
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
				encodeBinary(db)
				b.StopTimer()
			}
		})
	}
}

func BenchmarkDecodeBinary(b *testing.B) {
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

				data, _ := encodeBinary(db)
				b.StartTimer()
				decodeBinary(db, data)
				b.StopTimer()
			}
		})
	}
}

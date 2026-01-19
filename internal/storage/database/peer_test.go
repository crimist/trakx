package database

import (
	"testing"
	"time"

	"github.com/crimist/trakx/internal/stats"
	"github.com/crimist/trakx/internal/storage"
)

func TestPeerAdd(t *testing.T) {
	db, err := NewDatabase(Config{
		InitalSize:         1,
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Minute,
		Collector:          stats.NewCollectors(false, false, 0),
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
	db, err := NewDatabase(Config{
		InitalSize:         1,
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Minute,
		Collector:          stats.NewCollectors(false, false, 0),
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

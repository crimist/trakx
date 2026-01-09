package database

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
)

func TestSnapshotRoundTrip(t *testing.T) {
	db, err := NewDatabase(Config{
		InitalSize:         1,
		PersistanceAddress: "",
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

	var buf bytes.Buffer
	if err := db.Snapshot(&buf); err != nil {
		t.Fatal("Snapshot threw error: ", err)
	}

	if !bytes.HasPrefix(buf.Bytes(), []byte(snapshotMagic)) {
		t.Fatalf("snapshot missing magic header %q", snapshotMagic)
	}

	oldtorrents := db.torrents
	db, err = NewDatabase(Config{
		InitalSize:         1,
		PersistanceAddress: "",
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Minute,
		Collector:          stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}
	if err := db.Restore(&buf); err != nil {
		t.Fatal("Restore threw error: ", err)
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

func TestRestoreLegacyBinary(t *testing.T) {
	db, err := NewDatabase(Config{
		InitalSize:         1,
		PersistanceAddress: "",
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

	data, err := encodeBinary(db)
	if err != nil {
		t.Fatal("encodeBinary threw error: ", err)
	}
	oldtorrents := db.torrents

	db, err = NewDatabase(Config{
		InitalSize:         1,
		PersistanceAddress: "",
		EvictionFrequency:  1 * time.Minute,
		ExpirationTime:     1 * time.Minute,
		Collector:          stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}
	if err := db.Restore(bytes.NewReader(data)); err != nil {
		t.Fatal("Restore threw error: ", err)
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

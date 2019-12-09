package gomap

import (
	"testing"

	"github.com/crimist/trakx/tracker/shared"
)

const (
	benchHashes     = 150_000
	benchPeers      = 3
	benchPeersTotal = benchHashes * benchPeers
)

func TestCheck(t *testing.T) {
	var db Memory
	if db.Check() != false {
		t.Error("check() on empty db returned true")
	}
}

func TestTrim(t *testing.T) {
	var c shared.Config
	c.Database.Peer.Timeout = 0

	db := dbWithHashes(150_000)
	db.conf = &c

	db.trim()

	db, _ = dbWithPeers(200_000)
	db.conf = &c

	db.trim()
}

func BenchmarkTrim(b *testing.B) {
	cfg := new(shared.Config)
	cfg.Database.Peer.Timeout = -1

	b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		peerdb := dbWithHashesAndPeers(benchHashes, benchPeers)
		peerdb.conf = cfg

		b.StartTimer()
		peerdb.trim()
		b.StopTimer()
	}
}

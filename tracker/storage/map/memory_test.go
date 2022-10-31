package gomap

import (
	"testing"
	"time"

	"github.com/crimist/trakx/tracker/config"
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
	config.Conf.DB.Expiry = 0

	db := dbWithHashes(150_000)
	db.trim()

	db, _ = dbWithPeers(200_000)
	db.trim()
}

func BenchmarkTrim(b *testing.B) {
	config.Conf.DB.Expiry = -1 * time.Second

	b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		peerdb := dbWithHashesAndPeers(benchHashes, benchPeers)

		b.StartTimer()
		peerdb.trim()
		b.StopTimer()
	}
}

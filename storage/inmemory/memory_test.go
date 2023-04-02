package inmemory

import (
	"testing"
	"time"

	"github.com/crimist/trakx/config"
)

const (
	benchHashes     = 150_000
	benchPeers      = 3
	benchPeersTotal = benchHashes * benchPeers
)

func TestCheck(t *testing.T) {
	var db InMemory
	if db.Check() != false {
		t.Error("check() on empty db returned true")
	}
}

func TestTrim(t *testing.T) {
	config.Config.DB.Expiry = 0

	db := dbWithHashes(150_000)
	db.trim()

	db, _ = dbWithPeers(200_000)
	db.trim()
}

func BenchmarkTrim(b *testing.B) {
	config.Config.DB.Expiry = -1 * time.Second

	b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		peerdb := dbWithHashesAndPeers(benchHashes, benchPeers)

		b.StartTimer()
		peerdb.trim()
		b.StopTimer()
	}
}

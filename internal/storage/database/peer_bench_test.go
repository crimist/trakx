package database

import (
	"math/rand"
	"testing"
	"time"

	"github.com/crimist/trakx/internal/stats"
	"github.com/crimist/trakx/internal/storage"
)

const (
	benchPeerCount    = 1 << 15
	benchPeerMask     = benchPeerCount - 1
	benchPeerShift    = 15
	benchTorrentCount = 1 << 8
	benchTorrentMask  = benchTorrentCount - 1
)

func newBenchDB(b *testing.B) *Database {
	b.Helper()
	db, err := NewDatabase(Config{
		InitalSize:        1,
		EvictionFrequency: 1 * time.Minute,
		ExpirationTime:    1 * time.Minute,
		Collector:         stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		b.Fatal("Failed to create database")
	}
	return db
}

func makePeerIDs(n int, seed int64) []storage.PeerID {
	rnd := rand.New(rand.NewSource(seed))
	ids := make([]storage.PeerID, n)
	for i := range ids {
		rnd.Read(ids[i][:])
	}
	return ids
}

func makeHashes(n int, seed int64) []storage.Hash {
	rnd := rand.New(rand.NewSource(seed))
	hashes := make([]storage.Hash, n)
	for i := range hashes {
		rnd.Read(hashes[i][:])
	}
	return hashes
}

func prefillSingleTorrent(db *Database, ids []storage.PeerID) {
	for i, id := range ids {
		complete := (i & 1) == 0
		db.PeerAdd(testTorrentHash1, id, testPeerIP, 1234, complete)
	}
}

func prefillMultiTorrent(db *Database, ids []storage.PeerID, hashes []storage.Hash) {
	for i, id := range ids {
		hash := hashes[i&benchTorrentMask]
		complete := (i & 1) == 0
		db.PeerAdd(hash, id, testPeerIP, 1234, complete)
	}
}

func BenchmarkPeerAddSingle(b *testing.B) {
	ids := makePeerIDs(benchPeerCount, 1)

	b.Run("NewPeer", func(b *testing.B) {
		db := newBenchDB(b)
		idx := 0
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			if idx == len(ids) {
				b.StopTimer()
				db = newBenchDB(b)
				idx = 0
				b.StartTimer()
			}
			id := ids[idx]
			complete := (idx & 1) == 0
			db.PeerAdd(testTorrentHash1, id, testPeerIP, 1234, complete)
			idx++
		}
	})

	b.Run("ExistingPeer", func(b *testing.B) {
		db := newBenchDB(b)
		prefillSingleTorrent(db, ids)
		idx := 0
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			id := ids[idx&benchPeerMask]
			cycle := idx >> benchPeerShift
			complete := (cycle & 1) == 0
			db.PeerAdd(testTorrentHash1, id, testPeerIP, 1234, complete)
			idx++
		}
	})
}

func BenchmarkPeerAddSingleParallell(b *testing.B) {
	ids := makePeerIDs(benchPeerCount, 2)

	b.Run("ExistingPeer", func(b *testing.B) {
		db := newBenchDB(b)
		prefillSingleTorrent(db, ids)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			idx := 0
			for pb.Next() {
				id := ids[idx&benchPeerMask]
				cycle := idx >> benchPeerShift
				complete := (cycle & 1) == 0
				db.PeerAdd(testTorrentHash1, id, testPeerIP, 1234, complete)
				idx++
			}
		})
	})

	b.Run("MixedPeerIDs", func(b *testing.B) {
		db := newBenchDB(b)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			idx := 0
			for pb.Next() {
				id := ids[idx&benchPeerMask]
				complete := (idx & 1) == 0
				db.PeerAdd(testTorrentHash1, id, testPeerIP, 1234, complete)
				idx++
			}
		})
	})
}

func BenchmarkPeerAddMulti(b *testing.B) {
	ids := makePeerIDs(benchPeerCount, 3)
	hashes := makeHashes(benchTorrentCount, 4)

	b.Run("NewPeer", func(b *testing.B) {
		db := newBenchDB(b)
		idx := 0
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			if idx == len(ids) {
				b.StopTimer()
				db = newBenchDB(b)
				idx = 0
				b.StartTimer()
			}
			id := ids[idx]
			hash := hashes[idx&benchTorrentMask]
			complete := (idx & 1) == 0
			db.PeerAdd(hash, id, testPeerIP, 1234, complete)
			idx++
		}
	})

	b.Run("ExistingPeer", func(b *testing.B) {
		db := newBenchDB(b)
		prefillMultiTorrent(db, ids, hashes)
		idx := 0
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			id := ids[idx&benchPeerMask]
			hash := hashes[idx&benchTorrentMask]
			cycle := idx >> benchPeerShift
			complete := (cycle & 1) == 0
			db.PeerAdd(hash, id, testPeerIP, 1234, complete)
			idx++
		}
	})
}

func BenchmarkPeerAddMultiParallell(b *testing.B) {
	ids := makePeerIDs(benchPeerCount, 5)
	hashes := makeHashes(benchTorrentCount, 6)

	b.Run("ExistingPeer", func(b *testing.B) {
		db := newBenchDB(b)
		prefillMultiTorrent(db, ids, hashes)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			idx := 0
			for pb.Next() {
				id := ids[idx&benchPeerMask]
				hash := hashes[idx&benchTorrentMask]
				cycle := idx >> benchPeerShift
				complete := (cycle & 1) == 0
				db.PeerAdd(hash, id, testPeerIP, 1234, complete)
				idx++
			}
		})
	})

	b.Run("MixedPeerIDs", func(b *testing.B) {
		db := newBenchDB(b)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			idx := 0
			for pb.Next() {
				id := ids[idx&benchPeerMask]
				hash := hashes[idx&benchTorrentMask]
				complete := (idx & 1) == 0
				db.PeerAdd(hash, id, testPeerIP, 1234, complete)
				idx++
			}
		})
	})
}

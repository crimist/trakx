package gomap

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

func dbWithHashesAndPeers(hashes, peers int) *Memory {
	var err error
	var db Memory
	db.conf = new(shared.Config)

	storage.Expvar.Seeds = int64(hashes * peers)

	db.make()
	db.Expvar()
	if db.conf.Logger, err = zap.NewDevelopment(); err != nil {
		panic(err)
	}

	peerid := storage.PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	peer := storage.Peer{
		Complete: true,
		IP:       storage.PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}

	var h storage.Hash
	for i := 0; i < hashes; i++ {
		hash := make([]byte, 20)
		rand.Read(hash[:])
		copy(h[:], hash)

		for i := 0; i < peers; i++ {
			rand.Read(peerid[:])
			db.Save(&peer, h, peerid)
		}
	}

	return &db
}

func dbWithHashes(count int) *Memory {
	var db Memory
	db.make()
	db.Expvar()

	peerid := storage.PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	peer := storage.Peer{
		Complete: true,
		IP:       storage.PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}

	var h storage.Hash
	for i := 0; i < count; i++ {
		hash := make([]byte, 20)
		rand.Read(hash)
		copy(h[:], hash)

		db.Save(&peer, h, peerid)
	}

	return &db
}

func dbWithPeers(count int) (*Memory, storage.Hash) {
	var db Memory
	db.make()
	db.Expvar()

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := storage.Hash(bytes)
	peer := storage.Peer{
		Complete: true,
		IP:       storage.PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}

	var p storage.PeerID
	for i := 0; i < count; i++ {
		peerid := make([]byte, 20)
		rand.Read(peerid)
		copy(p[:], peerid)

		db.Save(&peer, hash, p)
	}

	return &db, hash
}

func TestPeerList(t *testing.T) {
	var db Memory
	db.make()

	xd := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := storage.Hash(xd)

	d := bencoding.NewDict()
	peerlist := db.PeerListBytes(hash, 0)
	d.String("peers", *(*string)(unsafe.Pointer(&peerlist)))
}

func benchmarkHashes(b *testing.B, count int) {
	db := dbWithHashes(count)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		db.Hashes()
	}
}

func BenchmarkHashes0(b *testing.B)      { benchmarkHashes(b, 0) }
func BenchmarkHashes5000(b *testing.B)   { benchmarkHashes(b, 5000) }
func BenchmarkHashes50000(b *testing.B)  { benchmarkHashes(b, 50000) }
func BenchmarkHashes500000(b *testing.B) { benchmarkHashes(b, 500000) }

// more/less peers doesn't change performance
const numPeers = 1000

func benchmarkPeerList(b *testing.B, cap int) {
	db, hash := dbWithPeers(numPeers)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		db.PeerList(hash, cap, false)
	}
}

func BenchmarkPeerList50(b *testing.B)  { benchmarkPeerList(b, 50) }
func BenchmarkPeerList100(b *testing.B) { benchmarkPeerList(b, 100) }
func BenchmarkPeerList200(b *testing.B) { benchmarkPeerList(b, 200) }
func BenchmarkPeerList400(b *testing.B) { benchmarkPeerList(b, 400) }

func benchmarkPeerListNopeerid(b *testing.B, cap int) {
	db, hash := dbWithPeers(numPeers)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		db.PeerList(hash, cap, true)
	}
}

func BenchmarkPeerListNopeerid50(b *testing.B)  { benchmarkPeerListNopeerid(b, 50) }
func BenchmarkPeerListNopeerid100(b *testing.B) { benchmarkPeerListNopeerid(b, 100) }
func BenchmarkPeerListNopeerid200(b *testing.B) { benchmarkPeerListNopeerid(b, 200) }
func BenchmarkPeerListNopeerid400(b *testing.B) { benchmarkPeerListNopeerid(b, 400) }

func benchmarkPeerListBytes(b *testing.B, cap int) {
	db, hash := dbWithPeers(numPeers)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		db.PeerListBytes(hash, cap)
	}
}

func BenchmarkPeerListBytes50(b *testing.B)  { benchmarkPeerListBytes(b, 50) }
func BenchmarkPeerListBytes100(b *testing.B) { benchmarkPeerListBytes(b, 100) }
func BenchmarkPeerListBytes200(b *testing.B) { benchmarkPeerListBytes(b, 200) }
func BenchmarkPeerListBytes400(b *testing.B) { benchmarkPeerListBytes(b, 400) }

func benchmarkHashStats(b *testing.B, peers int) {
	db, hash := dbWithPeers(peers)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		db.HashStats(hash)
	}
}

func BenchmarkHashStats100(b *testing.B)  { benchmarkHashStats(b, 100) }
func BenchmarkHashStats1000(b *testing.B) { benchmarkHashStats(b, 1000) }

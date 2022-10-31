package gomap

import (
	"net/netip"
	"testing"
	"time"

	"github.com/crimist/trakx/tracker/storage"
)

var (
	testIP   = netip.AddrFrom4([4]byte{1, 2, 3, 4})
	testHash = storage.Hash([20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	testId   = storage.PeerID([20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
)

func TestSaveDrop(t *testing.T) {
	var db Memory
	db.make()
	db.SyncExpvars()

	peerWrite := storage.Peer{
		Complete: true,
		IP:       testIP,
		Port:     4321,
	}
	db.Save(peerWrite.IP, peerWrite.Port, peerWrite.Complete, testHash, testId)
	peerRead, ok := db.hashmap[testHash].Peers[testId]

	if !ok {
		t.Error("Failed to read peer from database map")
	}
	if peerRead.Complete != peerWrite.Complete {
		t.Errorf("Peer complete not equal %v:%v", peerRead.Complete, peerWrite.Complete)
	}
	if peerRead.IP != peerWrite.IP {
		t.Errorf("Peer IP not equal %v:%v", peerRead.IP, peerWrite.IP)
	}
	if peerRead.Port != peerWrite.Port {
		t.Errorf("Peer port not equal %v:%v", peerRead.Port, peerWrite.Port)
	}
	if peerRead.LastSeen != time.Now().Unix() {
		t.Errorf("Peer LastSeen not correct %v:%v", peerRead.LastSeen, time.Now().Unix())
	}

	db.Drop(testHash, testId)
	_, ok = db.hashmap[testHash].Peers[testId]

	if ok {
		t.Error("Failed top drop peer from database")
	}
}

func benchmarkSave(b *testing.B, db *Memory, peer storage.Peer, hash storage.Hash, peerid storage.PeerID) {
	for n := 0; n < b.N; n++ {
		db.Save(peer.IP, peer.Port, peer.Complete, hash, peerid)
	}
}

func BenchmarkSave(b *testing.B) {
	var db Memory
	db.make()
	db.SyncExpvars()

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := storage.Hash(bytes)
	peerid := storage.PeerID(bytes)
	peer := storage.Peer{
		Complete: true,
		IP:       testIP,
		Port:     4321,
		LastSeen: 1234567890,
	}

	b.ResetTimer()
	benchmarkSave(b, &db, peer, hash, peerid)
}

func benchmarkDrop(b *testing.B, db *Memory, hash storage.Hash, peerid storage.PeerID) {
	for n := 0; n < b.N; n++ {
		db.Drop(hash, peerid)
	}
}

func BenchmarkDrop(b *testing.B) {
	var db Memory
	db.make()
	db.SyncExpvars()

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := storage.Hash(bytes)
	peerid := storage.PeerID(bytes)

	b.ResetTimer()
	benchmarkDrop(b, &db, hash, peerid)
}

func benchmarkSaveDrop(b *testing.B, db *Memory, peer storage.Peer, hash storage.Hash, peerid storage.PeerID) {
	for n := 0; n < b.N; n++ {
		db.Save(peer.IP, peer.Port, peer.Complete, hash, peerid)
		db.Drop(hash, peerid)
	}
}

func BenchmarkSaveDrop(b *testing.B) {
	var db Memory
	db.make()
	db.SyncExpvars()

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := storage.Hash(bytes)
	peerid := storage.PeerID(bytes)
	peer := storage.Peer{
		Complete: true,
		IP:       testIP,
		Port:     4321,
		LastSeen: 1234567890,
	}

	b.ResetTimer()
	benchmarkSaveDrop(b, &db, peer, hash, peerid)
}

func benchmarkSaveDropParallel(b *testing.B, routines int) {
	var db Memory
	db.make()
	db.SyncExpvars()

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := storage.Hash(bytes)
	peerid := storage.PeerID(bytes)
	peer := storage.Peer{
		Complete: true,
		IP:       testIP,
		Port:     4321,
		LastSeen: 1234567890,
	}

	b.SetParallelism(routines)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db.Save(peer.IP, peer.Port, peer.Complete, hash, peerid)
			db.Drop(hash, peerid)
		}
	})
}

func BenchmarkSaveDropParallel16(b *testing.B)  { benchmarkSaveDropParallel(b, 16) }
func BenchmarkSaveDropParallel32(b *testing.B)  { benchmarkSaveDropParallel(b, 32) }
func BenchmarkSaveDropParallel64(b *testing.B)  { benchmarkSaveDropParallel(b, 64) }
func BenchmarkSaveDropParallel128(b *testing.B) { benchmarkSaveDropParallel(b, 128) }
func BenchmarkSaveDropParallel256(b *testing.B) { benchmarkSaveDropParallel(b, 256) }
func BenchmarkSaveDropParallel512(b *testing.B) { benchmarkSaveDropParallel(b, 512) }

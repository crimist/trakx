package inmemory

import (
	"bytes"
	"testing"
)

func TestSaveDrop(t *testing.T) {
	var db PeerDatabase
	db.make()
	InitExpvar(&db)

	exbytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := Hash(exbytes)
	peerid := PeerID(exbytes)

	savePeer := Peer{
		Complete: true,
		IP:       PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}
	db.Save(&savePeer, &hash, &peerid)

	getPeer, ok := db.hashmap[hash].peers[peerid]
	if !ok {
		t.Error("Getting peer not ok")
	}
	if getPeer.Complete != savePeer.Complete {
		t.Error("Complete not equal")
	}
	if !bytes.Equal(getPeer.IP[:], savePeer.IP[:]) {
		t.Error("IP not equal")
	}
	if getPeer.Port != savePeer.Port {
		t.Error("Port not equal")
	}
	if getPeer.LastSeen != savePeer.LastSeen {
		t.Error("LastSeen not equal")
	}

	db.Drop(&savePeer, &hash, &peerid)
}

func benchmarkSave(b *testing.B, db *PeerDatabase, peer Peer, hash Hash, peerid PeerID) {
	for n := 0; n < b.N; n++ {
		db.Save(&peer, &hash, &peerid)
	}
}

func BenchmarkSave(b *testing.B) {
	var db PeerDatabase
	db.make()
	InitExpvar(&db)

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := Hash(bytes)
	peerid := PeerID(bytes)
	peer := Peer{
		Complete: true,
		IP:       PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}

	b.ResetTimer()
	benchmarkSave(b, &db, peer, hash, peerid)
}

func benchmarkDrop(b *testing.B, db *PeerDatabase, peer Peer, hash Hash, peerid PeerID) {
	for n := 0; n < b.N; n++ {
		db.Drop(&peer, &hash, &peerid)
	}
}

func BenchmarkDrop(b *testing.B) {
	var db PeerDatabase
	db.make()
	InitExpvar(&db)

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := Hash(bytes)
	peerid := PeerID(bytes)
	peer := Peer{
		Complete: true,
		IP:       PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}

	b.ResetTimer()
	benchmarkDrop(b, &db, peer, hash, peerid)
}

func benchmarkSaveDrop(b *testing.B, db *PeerDatabase, peer Peer, hash Hash, peerid PeerID) {
	for n := 0; n < b.N; n++ {
		db.Save(&peer, &hash, &peerid)
		db.Drop(&peer, &hash, &peerid)
	}
}

func BenchmarkSaveDrop(b *testing.B) {
	var db PeerDatabase
	db.make()
	InitExpvar(&db)

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := Hash(bytes)
	peerid := PeerID(bytes)
	peer := Peer{
		Complete: true,
		IP:       PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}

	b.ResetTimer()
	benchmarkSaveDrop(b, &db, peer, hash, peerid)
}

func benchmarkSaveDropParallel(b *testing.B, routines int) {
	var db PeerDatabase
	db.make()
	InitExpvar(&db)

	bytes := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := Hash(bytes)
	peerid := PeerID(bytes)
	peer := Peer{
		Complete: true,
		IP:       PeerIP{1, 2, 3, 4},
		Port:     4321,
		LastSeen: 1234567890,
	}

	b.SetParallelism(routines)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db.Save(&peer, &hash, &peerid)
			db.Drop(&peer, &hash, &peerid)
		}
	})
}

func BenchmarkSaveDropParallel16(b *testing.B)  { benchmarkSaveDropParallel(b, 16) }
func BenchmarkSaveDropParallel32(b *testing.B)  { benchmarkSaveDropParallel(b, 32) }
func BenchmarkSaveDropParallel64(b *testing.B)  { benchmarkSaveDropParallel(b, 64) }
func BenchmarkSaveDropParallel128(b *testing.B) { benchmarkSaveDropParallel(b, 128) }
func BenchmarkSaveDropParallel256(b *testing.B) { benchmarkSaveDropParallel(b, 256) }
func BenchmarkSaveDropParallel512(b *testing.B) { benchmarkSaveDropParallel(b, 512) }

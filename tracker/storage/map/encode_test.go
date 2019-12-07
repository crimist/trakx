package gomap

import (
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

func TestEncodeDecode(t *testing.T) {
	db := dbWithHashesAndPeers(1000, 5)

	data, err := db.encode()
	if err != nil {
		t.Fatal(err)
	}

	peers, hashes, err := db.decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if peers != 5000 {
		t.Fatal("peers != 5000")
	}
	if hashes != 1000 {
		t.Fatal("hashes != 1000")
	}
}

func TestEncodeBinary(t *testing.T) {
	var err error
	var db Memory
	db.conf = new(shared.Config)

	db.make()
	db.Expvar()
	if db.conf.Logger, err = zap.NewDevelopment(); err != nil {
		panic(err)
	}

	hash := storage.Hash{0x48, 0x61, 0x73, 0x68, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	peerid := storage.PeerID{0x49, 0x44, 0x49, 0x44, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	peer := storage.Peer{
		Complete: true,
		IP:       storage.PeerIP{0x49, 0x50, 0x44, 0x52}, // IPDR
		Port:     0x4f50,                                 // PO
		LastSeen: 0x4e4545535453414c,                     // LASTSEEN
	}
	db.Save(&peer, &hash, &peerid)

	data, _ := db.encodeBinary()

	db = Memory{}

	db.decodeBinary(data)
	submap := db.hashmap[hash]
	dbpeer := submap.peers[peerid]

	if !reflect.DeepEqual(*dbpeer, peer) {
		t.Fatal("Not equal!", spew.Sdump(peer, *dbpeer))
	}

	return
}

const (
	benchHashes = 150_000
	benchPeers  = 3
)

func BenchmarkEncode(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.encode()
	}
}

func BenchmarkEncodeBinary(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.encodeBinary()
	}
}

func BenchmarkDecode(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)
	buff, err := db.encode()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.decode(buff)
	}
}

func BenchmarkDecodeBinary(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)
	buff, err := db.encode()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.decodeBinary(buff)
	}
}

func BenchmarkEncodeMemuse(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		peerdb := dbWithHashesAndPeers(benchHashes, benchPeers)

		var start runtime.MemStats
		runtime.ReadMemStats(&start)

		b.StartTimer()
		encoded, _ := peerdb.encode()
		b.StopTimer()

		var end runtime.MemStats
		runtime.ReadMemStats(&end)

		b.Logf("Trim: %dMB using %dMB with %d GC cycles", len(encoded)/1024/1024, (end.HeapAlloc-start.HeapAlloc)/1024/1024, end.NumGC-start.NumGC)
		debug.FreeOSMemory()
	}
}

func BenchmarkEncodeBinaryMemuse(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()

	gcp := debug.SetGCPercent(-1)

	for i := 0; i < b.N; i++ {
		peerdb := dbWithHashesAndPeers(benchHashes, benchPeers)

		var start runtime.MemStats
		runtime.ReadMemStats(&start)

		b.StartTimer()
		encoded, _ := peerdb.encodeBinary()
		b.StopTimer()

		var end runtime.MemStats
		runtime.ReadMemStats(&end)

		b.Logf("Trim: %dMB using %dMB with %d GC cycles", len(encoded)/1024/1024, (end.HeapAlloc-start.HeapAlloc)/1024/1024, end.NumGC-start.NumGC)
		debug.FreeOSMemory()
	}

	debug.SetGCPercent(gcp)
}

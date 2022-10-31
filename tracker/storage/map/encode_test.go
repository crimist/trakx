package gomap

import (
	"encoding/hex"
	"net/netip"
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
)

// TODO: refactor these, deepequal the whole database (or at least the hashmap)

func TestEncodeBinary(t *testing.T) {
	config.Conf.LogLevel = "debug"
	err := config.Conf.Update()
	if err != nil {
		t.Error(err)
	}
	var db Memory

	db.make()
	db.SyncExpvars()

	hash := storage.Hash{0x48, 0x61, 0x73, 0x68, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	peerid := storage.PeerID{0x49, 0x44, 0x49, 0x44, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	peer := storage.Peer{
		Complete: true,
		IP:       netip.MustParseAddr("127.0.0.1"),
		Port:     0x4f50,
		LastSeen: time.Now().Unix(),
	}
	db.Save(peer.IP, peer.Port, peer.Complete, hash, peerid)

	data, _ := db.encodeBinary()

	db = Memory{}
	if _, _, err := db.decodeBinary(data); err != nil {
		t.Fatal("decodeBinary threw error: ", err)
	}
	submap := db.hashmap[hash]
	dbpeer := submap.peers[peerid]

	if !reflect.DeepEqual(*dbpeer, peer) {
		t.Fatal("Not equal!\n" + hex.Dump(data) + spew.Sdump(peer, *dbpeer))
	}
}

// encode benches

func BenchmarkEncodeBinary(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.encodeBinary()
	}
}

// decode benches

func BenchmarkDecodeBinary(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)
	buff, err := db.encodeBinary()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.decodeBinary(buff)
	}
}

// memuse benches

type encoder func() ([]byte, error)

func benchmarkMemuse(function encoder, b *testing.B) {
	b.StopTimer()
	b.ResetTimer()

	gcp := debug.SetGCPercent(-1)

	for i := 0; i < b.N; i++ {
		var start, end runtime.MemStats
		runtime.ReadMemStats(&start)

		b.StartTimer()
		encoded, _ := function()
		b.StopTimer()

		runtime.ReadMemStats(&end)

		b.Logf("Encode: %dMB using %dMB with %d GC cycles", len(encoded)/1024/1024, (end.HeapAlloc-start.HeapAlloc)/1024/1024, end.NumGC-start.NumGC)
		debug.FreeOSMemory()
	}

	debug.SetGCPercent(gcp)
}

func BenchmarkEncodeBinaryMemuse(b *testing.B) {
	peerdb := dbWithHashesAndPeers(benchHashes, benchPeers)
	benchmarkMemuse(peerdb.encodeBinary, b)
}

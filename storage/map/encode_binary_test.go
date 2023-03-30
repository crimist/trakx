package gomap

import (
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/storage"
)

func TestEncodeDecodeBinary(t *testing.T) {
	var db Memory
	db.make()
	pools.Initialize(10)

	hash := storage.Hash{0x48, 0x61, 0x73, 0x68, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	peerid := storage.PeerID{0x49, 0x44, 0x49, 0x44, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	peer := storage.Peer{
		Complete: true,
		IP:       netip.MustParseAddr("127.0.0.1"),
		Port:     0x4f50,
		LastSeen: time.Now().Unix(),
	}
	db.Save(peer.IP, peer.Port, peer.Complete, hash, peerid)

	oldhahmap := db.hashmap
	data, err := db.encodeBinary()
	if err != nil {
		t.Fatal("encodeBinary threw error: ", err)
	}
	db = Memory{}
	if _, _, err := db.decodeBinary(data); err != nil {
		t.Fatal("decodeBinary threw error: ", err)
	}

	if _, ok := db.hashmap[hash]; !ok {
		t.Fatal("hashmap not equal, missing hash entry")
	}
	if oldhahmap[hash].Complete != db.hashmap[hash].Complete {
		t.Fatalf("Complete not equal: should %v, got %v", oldhahmap[hash].Complete, db.hashmap[hash].Complete)
	}
	if oldhahmap[hash].Incomplete != db.hashmap[hash].Incomplete {
		t.Fatalf("Incomplete not equal: should %v, got %v", oldhahmap[hash].Incomplete, db.hashmap[hash].Incomplete)
	}
	if !reflect.DeepEqual(oldhahmap[hash].Peers, db.hashmap[hash].Peers) {
		t.Fatalf("Peer not equal: should %v, got %v", oldhahmap[hash].Peers, db.hashmap[hash].Peers)
	}
}

// binary benchmarks

func BenchmarkEncodeBinary(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.encodeBinary()
	}
}

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

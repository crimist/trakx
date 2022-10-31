package gomap

import (
	"encoding/hex"
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/davecgh/go-spew/spew"
)

func TestEncodeDecodeGob(t *testing.T) {
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

	data, err := db.encodeGob()
	if err != nil {
		t.Fatal("encodeGob threw error: ", err)
	}
	db = Memory{}
	if err := db.decodeGob(data); err != nil {
		t.Fatal("decodeGob threw error: ", err)
	}

	submap := db.hashmap[hash]
	dbpeer := submap.Peers[peerid]
	if !reflect.DeepEqual(*dbpeer, peer) {
		t.Fatal("Not equal!\n" + hex.Dump(data) + spew.Sdump(peer, *dbpeer))
	}
}

func BenchmarkEncodeGob(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.encodeGob()
	}
}

func BenchmarkDecodeGob(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)
	buff, err := db.encodeGob()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.decodeGob(buff)
	}
}

func BenchmarkEncodeGobMemuse(b *testing.B) {
	db := dbWithHashesAndPeers(benchHashes, benchPeers)
	benchmarkMemuse(db.encodeGob, b)
}

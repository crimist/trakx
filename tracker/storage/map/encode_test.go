package gomap

import (
	"runtime"
	"runtime/debug"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	db := dbWithHashesAndPeers(1000, 5)

	data, err := db.encode()
	if err != nil {
		t.Fatal(err)
	}

	err = db.decode(data)
	if err != nil {
		t.Fatal(err)
	}
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

func BenchmarkEncodeMemuse(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		peerdb := dbWithHashesAndPeers(benchHashes, benchPeers)

		var start runtime.MemStats
		runtime.ReadMemStats(&start)

		b.StartTimer()
		x, _ := peerdb.encode()
		b.StopTimer()

		var end runtime.MemStats
		runtime.ReadMemStats(&end)

		b.Logf("Trim: %dMB using %dMB with %d GC cycles", len(x)/1024/1024, (end.HeapAlloc-start.HeapAlloc)/1024/1024, end.NumGC-start.NumGC)
		debug.FreeOSMemory()
	}
}

package storage

import "testing"

func BenchmarkPeerChanMemCost(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var pc peerChan
		pc.create()
	}
}

func BenchmarkPeerChanGet(b *testing.B) {
	var pc peerChan
	pc.create()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.Get()
	}
}

func BenchmarkPeerChanPut(b *testing.B) {
	var pc peerChan
	pc.create()
	var p = new(Peer)

	b.Skip("Invalid results as the benchmark exceeds the cap of the channel")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.Put(p)
	}
}

func benchmarkPeerChanGetPut(b *testing.B, in int) {
	var pc peerChan
	pc.create()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := pc.Get()
		if i%in == 0 {
			pc.Put(p)
		}
	}
}

func BenchmarkPeerChanGetPut1_2(b *testing.B)  { benchmarkPeerChanGetPut(b, 2) }
func BenchmarkPeerChanGetPut1_4(b *testing.B)  { benchmarkPeerChanGetPut(b, 4) }
func BenchmarkPeerChanGetPut1_10(b *testing.B) { benchmarkPeerChanGetPut(b, 10) }
func BenchmarkPeerChanGetPut1_20(b *testing.B) { benchmarkPeerChanGetPut(b, 20) }

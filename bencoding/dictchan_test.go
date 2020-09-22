package bencoding

import "testing"

func BenchmarkPeerChanMemCost(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var dc dictCh
		dc.Init()
	}
}

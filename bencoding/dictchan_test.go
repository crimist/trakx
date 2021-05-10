package bencoding

import "testing"

func BenchmarkDictChanInit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var dc dictChannel
		dc.Init()
	}
}

package http

import (
	"net"
	"testing"
)

func BenchmarkWriteErr(b *testing.B) {
	c, _ := net.Dial("udp", ":1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeErr(c, "benchmark_string_test")
	}
}

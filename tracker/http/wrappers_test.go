package http

import (
	"strings"
	"testing"

	"github.com/cbeuw/connutil"
)

var writeDataBenchStr = strings.Repeat("A", 200)

func BenchmarkWriteData(b *testing.B) {
	conn := connutil.Discard()
	data := writeDataBenchStr

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeData(conn, data)
	}
}

func BenchmarkWriteDataBytes(b *testing.B) {
	conn := connutil.Discard()
	data := []byte(writeDataBenchStr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeDataBytes(conn, data)
	}
}

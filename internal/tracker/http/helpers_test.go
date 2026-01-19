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
		writeSuccess(conn, data)
	}
}

//go:build go1.18
// +build go1.18

package http

import (
	"testing"
)

// Fuzzing is in beta, see https://go.dev/blog/fuzz-beta

func FuzzParse(f *testing.F) {
	f.Add([]byte("GET /test?param=1&param2=two&test=test%3Ftest HTTP/1.1 bla bla"), 60)
	f.Fuzz(func(t *testing.T, data []byte, length int) {
		_, err := parse(data, length)
		if err != nil {
			t.Skip()
		}
	})
}

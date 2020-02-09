// +build heroku

package http

import (
	"bytes"
	"testing"
)

func TestForwarded(t *testing.T) {
	var cases = []struct {
		name string
		data []byte
		ip   []byte
	}{
		{"single", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 1.1.1.1\r\n\r\n"), []byte("1.1.1.1")},
		{"multi", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 1.1.1.1, 2.2.2.2\r\n\r\n"), []byte("2.2.2.2")},
		{"empty", []byte("GET / HTTP1.1\r\nX-Forwarded-For:\r\n\r\n"), nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, ip := getForwarded(c.data); !bytes.Equal(ip, c.ip) {
				t.Errorf("Bad ip '%s' should be '%s'", ip, c.ip)
			}
		})
	}
}

func BenchmarkForwarded(b *testing.B) {
	req := []byte("GET / HTTP/1.1\r\nX-Forwarded-For: 1.2.3.4\r\n\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ip := getForwarded(req)
		_ = ip
	}
}

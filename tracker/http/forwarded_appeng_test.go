//go:build heroku
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
		{"ipv4_single", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 1.1.1.1\r\n\r\n"), []byte("1.1.1.1")},
		{"ipv4_single", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 255.255.255.255\r\n\r\n"), []byte("255.255.255.255")},
		{"ipv4_multi", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 1.1.1.1, 2.2.2.2\r\n\r\n"), []byte("2.2.2.2")},
		{"ipv6_single_expanded", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 1111:2222:3333:4444:5555:6666:7777:8888\r\n\r\n"), []byte("1111:2222:3333:4444:5555:6666:7777:8888")},
		{"ipv6_single_small", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 2001:db8::2:1.\r\n\r\n"), []byte("2001:db8::2:1.")},
		{"ipv6_multi", []byte("GET / HTTP1.1\r\nX-Forwarded-For: 1111:2222:3333:4444:5555:6666:7777:8888, 9999:AAAA:BBBB:CCCC:DDDD:EEEE:FFFF:0000\r\n\r\n"), []byte("9999:AAAA:BBBB:CCCC:DDDD:EEEE:FFFF:0000")},
		{"space", []byte("GET / HTTP1.1\r\nX-Forwarded-For: \r\n\r\n"), nil},
		{"empty", []byte("GET / HTTP1.1\r\nX-Forwarded-For:\r\n\r\n"), nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, ip := parseForwarded(c.data); !bytes.Equal(ip, c.ip) {
				t.Errorf("bad ip '%s'; want '%s'", ip, c.ip)
			}
		})
	}
}

func BenchmarkForwarded(b *testing.B) {
	req := []byte("GET / HTTP/1.1\r\nX-Forwarded-For: 1.2.3.4\r\n\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ip := parseForwarded(req)
		_ = ip
	}
}

package http

import (
	"net"
	"testing"
)

const data = "testing data"

func BenchmarkWriteData(b *testing.B) {
	server, client := net.Pipe()
	go func() {
		b := make([]byte, 1000)
		for {
			server.Read(b)
		}
	}()

	var x = data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeData(client, x)
	}
}

func BenchmarkWriteDataBytes(b *testing.B) {
	server, client := net.Pipe()
	go func() {
		b := make([]byte, 1000)
		for {
			server.Read(b)
		}
	}()

	var x = []byte(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeDataBytes(client, x)
	}
}

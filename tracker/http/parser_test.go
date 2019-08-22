package http

import (
	"testing"
)

func TestParse(t *testing.T) {
	req := []byte("GET /test?param=1&test=test%3Ftest HTTP/1.1 bla bla")
	p, err := parse(req)

	if err != nil {
		t.Fatalf("Error when parsing: %v", err)
	}
	if len(p.Params[0]) == 0 {
		t.Fatal("Params not found")
	}
	for _, param := range p.Params {
		switch param {
		case "param=1":
		case "test=test?test":
		default:
			t.Fatalf("Incorrect params: %v", p.Params)
		}
	}
	if p.Path != "/test" {
		t.Fatal("Incorrect path")
	}
	if p.Method != "GET" {
		t.Fatalf("Incorrect method")
	}
}

func BenchmarkParse(b *testing.B) {
	req := []byte("GET /test HTTP/1.1 XXXXXXXXXXXXXXXX")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parse(req)
	}
}

func BenchmarkParseParams(b *testing.B) {
	req := []byte("GET /test?param=1?key=value HTTP/1.1 XXXXXXXXXXXXXXXX")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parse(req)
	}
}

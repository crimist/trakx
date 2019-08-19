package http

import "testing"

func TestParse(t *testing.T) {
	req := []byte("GET /test?param=1 HTTP/1.1 bla bla")
	p := parse(&req, len(req))

	if len(p.Params[0]) == 0 {
		t.Fatalf("Params not found")
	}
	if p.Params[0] != "param=1" {
		t.Fatalf("Params incorrectly parsed")
	}
	if p.Path != "/test" {
		t.Fatalf("Incorrect path")
	}
}

func BenchmarkParse(b *testing.B) {
	req := []byte("GET /test HTTP/1.1 XXXXXXXXXXXXXXXX")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parse(&req, len(req))
	}
}

func BenchmarkParseParams(b *testing.B) {
	req := []byte("GET /test?param=1?key=value HTTP/1.1 XXXXXXXXXXXXXXXX")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parse(&req, len(req))
	}
}

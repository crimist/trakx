package http

import "testing"

func TestParse(t *testing.T) {
	req := []byte("GET /test?param=1 HTTP/1.1 bla bla")
	p, err := parse(req)

	if err != nil {
		t.Fatalf("Error when parsing: %v", err)
	}
	if len(p.Params[0]) == 0 {
		t.Fatal("Params not found")
	}
	if p.Params[0] != "param=1" {
		t.Fatal("Params incorrectly parsed")
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

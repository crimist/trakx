package http

import (
	"net/url"
	"strings"
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

const benchRequest = "GET /benchmark HTTP/1.1\r\nHEADER: VALUE\r\n\r\n"
const benchReqParams = "GET /benchmark?key0=val0&key1=val1&key2=val2&key3=val3 HTTP/1.1\r\nHEADER: VALUE\r\n\r\n"

func BenchmarkParse(b *testing.B) {
	req := []byte(benchRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := parse(req)
		_ = p
	}
}

func BenchmarkURLParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p, _ := url.Parse(benchRequest[4:strings.Index(benchRequest, " HTTP/")])
		_ = p
	}
}

func BenchmarkParseParams(b *testing.B) {
	req := []byte(benchReqParams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := parse(req)
		_ = p
	}
}

func BenchmarkURLParseParams(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parsed, _ := url.Parse(benchReqParams[4:strings.Index(benchReqParams, " HTTP/")])
		p := parsed.Query()
		_ = p
	}
}

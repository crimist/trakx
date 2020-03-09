package http

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	req := []byte("GET /test?param=1&param2=two&test=test%3Ftest HTTP/1.1 bla bla")
	p, _, err := parse(req, len(req))

	if err != nil {
		t.Fatalf("Error when parsing: %v", err)
	}
	if len(p.Params[0]) == 0 {
		t.Fatal("Params not found")
	}
	for _, param := range p.Params {
		switch param {
		case "":
		case "param=1":
		case "param2=two":
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

	req = []byte("GET /url?key=value HTTP/1.1")
	p, _, err = parse(req, len(req))
	if err != nil {
		t.Fatalf("Error when parsing: %v", err)
	}
	if p.Params[0] != "key=value" {
		t.Fatal("Bad params")
	}
	if p.Path != "/url" {
		t.Fatal("Incorrect path", p.Path)
	}
	if p.Method != "GET" {
		t.Fatalf("Incorrect method")
	}
}

const benchRequest = "GET /benchmark HTTP/1.1\r\nHEADER: VALUE\r\n\r\n"

func BenchmarkParse(b *testing.B) {
	req := []byte(benchRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _, _ := parse(req, len(req))
		_ = p
	}
}

func BenchmarkStdParse(b *testing.B) {
	req := benchRequest[4:strings.Index(benchRequest, " HTTP/")]

	for i := 0; i < b.N; i++ {
		p, _ := url.Parse(req)
		_ = p
	}
}

const benchReqParams = "GET /benchmark?key0=val0&key1=val1&key2=val2&key3=val3 HTTP/1.1\r\nHEADER: VALUE\r\n\r\n"

func BenchmarkParseParams(b *testing.B) {
	req := []byte(benchReqParams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _, _ := parse(req, len(req))
		_ = p
	}
}

func BenchmarkParseParamsBase64(b *testing.B) {
	req := make([]byte, base64.StdEncoding.EncodedLen(len(benchReqParams)))
	base64.StdEncoding.Encode(req, []byte(benchReqParams))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _, _ := parse(req, len(req))
		_ = p
	}
}

func BenchmarkStdParseParams(b *testing.B) {
	req := benchReqParams[4:strings.Index(benchReqParams, " HTTP/")]

	for i := 0; i < b.N; i++ {
		parsed, _ := url.Parse(req)
		p := parsed.Query()
		_ = p
	}
}

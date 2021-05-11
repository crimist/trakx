package http

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	req := []byte("GET /test?param=1&param2=two&test=test%3Ftest HTTP/1.1 bla bla")
	p, err := parse(req, len(req))

	if err != nil {
		t.Fatalf("Error when parsing: %v", err)
	}
	if len(p.Params[0]) == 0 {
		t.Fatal("Params not found")
	}
	for _, param := range p.Params {
		switch string(param) {
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
	p, err = parse(req, len(req))
	if err != nil {
		t.Fatalf("Error when parsing: %v", err)
	}
	if !bytes.Equal(p.Params[0], []byte("key=value")) {
		t.Fatal("Bad params")
	}
	if p.Path != "/url" {
		t.Fatal("Incorrect path", p.Path)
	}
	if p.Method != "GET" {
		t.Fatalf("Incorrect method")
	}
}

func TestUnescapeFast(t *testing.T) {
	var cases = []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{"no escapes", []byte("test"), []byte("test")},
		{"one escape", []byte("1%002"), []byte("1\x002")},
		{"only escape", []byte("%ff"), []byte("\xff")},
		{"real info_hash", []byte("info_hash=%06%d4%cc2%9a%d79%7c%b854%99A%d4%1d%2c%b3%10H%3b"), []byte("info_hash=\x06\xd4\xcc2\x9a\xd79\x7c\xb854\x99A\xd4\x1d\x2c\xb3\x10H\x3b")},
		{"multipe escapes", []byte("1%002%ba~L"), []byte("1\x002\xba~L")},
		{"invalid escapes", []byte("1%2"), nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			unescaped := unescapeFast(c.input)
			if !bytes.Equal(unescaped, c.expected) {
				t.Errorf("Bad unescape\nGot:\n%v\nExpected:\n'%v'\n", hex.Dump(unescaped), hex.Dump(c.expected))
			}
		})
	}
}

const benchRequest = "GET /benchmark HTTP/1.1\r\nHEADER: VALUE\r\n\r\n"

func BenchmarkParseBasic(b *testing.B) {
	req := []byte(benchRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := parse(req, len(req))
		_ = p
	}
}

func BenchmarkStdParseBasic(b *testing.B) {
	req := benchRequest[4:strings.Index(benchRequest, " HTTP/")]

	for i := 0; i < b.N; i++ {
		p, _ := url.Parse(req)
		_ = p
	}
}

const benchReqParams = "GET /benchmark?key0=val0&key1=val1%00%01%02%03%04%05%06%07%08%09%0A&key2=val2&key3=val3 HTTP/1.1\r\nHEADER: VALUE\r\n\r\n"

func BenchmarkParseParams(b *testing.B) {
	req := []byte(benchReqParams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := parse(req, len(req))
		_ = p
	}
}

func BenchmarkParseParamsBase64(b *testing.B) {
	req := make([]byte, base64.StdEncoding.EncodedLen(len(benchReqParams)))
	base64.StdEncoding.Encode(req, []byte(benchReqParams))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := parse(req, len(req))
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

const escapedParameter = "info_hash=%b4%8f%85%20%92%1e~%c7%c38%90%40gv%28%7b%e8%f9~L"

var escapedParameterBytes = []byte(escapedParameter)

func BenchmarkStdUnescape(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s, err := url.QueryUnescape(escapedParameter)
		_, _ = s, err
	}
}

func BenchmarkUnescape(b *testing.B) {
	for i := 0; i < b.N; i++ {
		unescapeFast(escapedParameterBytes)
	}
}

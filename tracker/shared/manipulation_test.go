package shared

import (
	"bytes"
	"strings"
	"testing"
)

func TestStringToBytes(t *testing.T) {
	var cases = []struct {
		name string
		data string
		size int
	}{
		{"small", "lmao", 4},
		{"big", "This CONTAINS test DATA!! :)", 28},
		{"non ascii", "\x01\x02\x03\x04\r\n", 6},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			converted := StringToBytes(c.data)
			if bytes.Compare([]byte(c.data), converted) != 0 {
				t.Fatal("Failed to convert data")
			}
			if c.size != len(converted) {
				t.Fatal("Invalid len()")
			}
			if c.size != cap(converted) {
				t.Fatal("Invalid cap()")
			}
		})
	}
}

func TestStringToBytesFast(t *testing.T) {
	var cases = []struct {
		name string
		data string
		size int
	}{
		{"small", "lmao", 4},
		{"big", "This CONTAINS test DATA!! :)", 28},
		{"non ascii", "\x01\x02\x03\x04\r\n", 6},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			converted := StringToBytesFast(&c.data)
			if bytes.Compare([]byte(c.data), converted) != 0 {
				t.Fatal("Failed to convert data")
			}
			if c.size != len(converted) {
				t.Fatal("Invalid len()")
			}
			if c.size != cap(converted) {
				t.Fatal("Invalid cap()")
			}
		})
	}
}

func TestSetSliceLen(t *testing.T) {
	var cases = []struct {
		name     string
		data     []byte
		size     int
		fullSize int
	}{
		{"short", []byte("data"), 2, 4},
		{"long", []byte("lots of stuff in here"), 4, 21},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			oldlen := SetSliceLen(&c.data, c.size)
			if len(c.data) != c.size {
				t.Fatal("Failed to set len")
			}

			SetSliceLen(&c.data, oldlen)
			if len(c.data) != c.fullSize {
				t.Fatal("Failed to restore len")
			}
		})
	}
}

var benchData = strings.Repeat("String", 1e4)
var benchByteData = bytes.Repeat([]byte("Bytes"), 1e4)

func BenchmarkStringToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StringToBytes(benchData)
	}
}

func BenchmarkStringToBytesFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StringToBytesFast(&benchData)
	}
}

func BenchmarkSetSliceLen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		oldlen := SetSliceLen(&benchByteData, 20)
		SetSliceLen(&benchByteData, oldlen)
	}
}

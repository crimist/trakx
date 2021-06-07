package bencoding

import (
	"strconv"
	"strings"
	"testing"
)

func TestEncodingStr(t *testing.T) {
	var cases = []struct {
		name     string
		input    string
		expected string
	}{
		{"short", "test", "4:test"},
		{"long", strings.Repeat("A", 1000), "1000:" + strings.Repeat("A", 1000)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := str(c.input)

			if out != c.expected {
				t.Logf("Bad announce\nGot:\n%v\nExpected:\n%v", out, c.expected)
			}
		})
	}
}

func TestEncodingInteger(t *testing.T) {
	var cases = []struct {
		name     string
		input    uint64
		expected string
	}{
		{"short", 0, "i0e"},
		{"long", ^uint64(0), "i" + strconv.FormatUint(^uint64(0), 10) + "e"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := integer(c.input)

			if out != c.expected {
				t.Logf("Bad announce\nGot:\n%v\nExpected:\n%v", out, c.expected)
			}
		})
	}
}

func TestEncodingList(t *testing.T) {
	var cases = []struct {
		name     string
		input    []string
		expected string
	}{
		{"short", []string{"hello", "world"}, "l5:hello5:worlde"},
		{"long", []string{"abc", "0123456789", "longer string :)", "yep", strings.Repeat("A", 1000)}, "l3:abc10:012345678916:longer string :)3:yep1000:" + strings.Repeat("A", 1000) + "e"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := list(c.input...)

			if out != c.expected {
				t.Logf("Bad announce\nGot:\n%v\nExpected:\n%v", out, c.expected)
			}
		})
	}
}

func BenchmarkEncodingStrShort(b *testing.B) { benchmarkEncodingStr(b, "test") }
func BenchmarkEncodingStrLong(b *testing.B)  { benchmarkEncodingStr(b, strings.Repeat("A", 1000)) }

func benchmarkEncodingStr(b *testing.B, s string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		str(s)
	}
}

func BenchmarkEncodingIntegerShort(b *testing.B) { benchmarkEncodingInteger(b, 0) }
func BenchmarkEncodingIntegerLong(b *testing.B)  { benchmarkEncodingInteger(b, ^uint64(0)) }

func benchmarkEncodingInteger(b *testing.B, i uint64) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		integer(i)
	}
}

func BenchmarkEncodingListShort(b *testing.B) { benchmarkEncodingList(b, []string{"hello", "world"}) }
func BenchmarkEncodingListLong(b *testing.B) {
	benchmarkEncodingList(b, []string{"abc", "0123456789", "longer string :)", "yep", strings.Repeat("A", 1000)})
}

func benchmarkEncodingList(b *testing.B, l []string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list(l...)
	}
}

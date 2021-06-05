package bencoding

import (
	"testing"
)

func TestDictionaryGetPut(t *testing.T) {
	d := GetDictionary()
	PutDictionary(d)

	if len(dictionaryChan) != 1 {
		t.Error("dictionaryChan len() != 1 after get & put")
	}
}

func BenchmarkDictionaryGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetDictionary()
	}
}

func BenchmarkDictionaryGetPut(b *testing.B) {
	// recycles same dict - should be 0 mem use

	for i := 0; i < b.N; i++ {
		d := GetDictionary()
		PutDictionary(d)
	}
}

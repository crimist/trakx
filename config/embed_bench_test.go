package config

import (
	"math/rand"
	"testing"
)

var filenames = [...]string{"/index.html", "/dmca.html"}

func randomFilename() string {
	return filenames[rand.Intn(len(filenames))]
}

func BenchmarkFSReadFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		data, _ := embeddedFS.ReadFile("embededd" + randomFilename())
		_ = data
	}
}

func BenchmarkEmbeddedCache(b *testing.B) {
	cache, err := GenerateEmbeddedCache()
	if err != nil {
		b.Errorf("failed to create cache: %v", err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		data := cache[randomFilename()]
		_ = data
	}
}

func BenchmarkSwitchRaw(b *testing.B) {
	indexData, err := embeddedFS.ReadFile("embedded/index.html")
	if err != nil {
		b.Error("failed to read embedded/index.html")
	}
	dmcaData, err := embeddedFS.ReadFile("embedded/dmca.html")
	if err != nil {
		b.Error("failed to read embedded/dmca.html")
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		switch randomFilename() {
		case "index.html":
			data := indexData
			_ = data
		case "dmca.html":
			data := dmcaData
			_ = data
		}
	}
}

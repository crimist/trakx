package gomap

import "testing"

func TestEncodeDecode(t *testing.T) {
	db := dbWithHashesAndPeers(1000, 5)

	data, err := db.encode()
	if err != nil {
		t.Fatal(err)
	}

	err = db.decode(data)
	if err != nil {
		t.Fatal(err)
	}
}

const (
	hashes = 80_000
	peers  = 3
)

func BenchmarkEncode(b *testing.B) {
	db := dbWithHashesAndPeers(hashes, peers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.encode()
	}
}

func BenchmarkDecode(b *testing.B) {
	db := dbWithHashesAndPeers(hashes, peers)
	buff, err := db.encode()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.decode(buff)
	}
}

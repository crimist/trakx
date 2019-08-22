package shared

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/gob"
	"encoding/hex"
	"io"
	"log"
	"testing"
)

func TestCheck(t *testing.T) {
	var db PeerDatabase
	if db.check() != false {
		t.Error("check() on empty db returned true")
	}
}

func TestTrim(t *testing.T) {
	var c Config
	c.Database.Peer.Timeout = 0

	db := dbWithHashes(1000000)
	db.conf = &c

	p, h := db.trim()
	t.Logf("Peers: %v Hashes: %v", p, h)

	db, _ = dbWithPeers(1000000)
	db.conf = &c

	p, h = db.trim()
	t.Logf("Peers: %v Hashes: %v", p, h)
}

func BenchmarkZip(b *testing.B) {
	db := dbWithHashesAndPeers(82000, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buff bytes.Buffer
		archive := zip.NewWriter(&buff)
		archive.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(out, flate.NoCompression) // flate nocomp is fastest
		})

		for hash, submap := range db.hashmap {
			writer, err := archive.Create(hex.EncodeToString(hash[:]))
			if err != nil {
				b.Error("Failed to create in archive", err, hash)
			}
			if err := gob.NewEncoder(writer).Encode(submap.peers); err != nil {
				b.Error("Failed to encode a peermap", err, hash)
			}
		}

		if err := archive.Close(); err != nil {
			b.Error("Failed to close archive", err)
		}

		b.Log(len(buff.Bytes()))
	}
}

func BenchmarkTar(b *testing.B) {
	db := dbWithHashesAndPeers(82000, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buff bytes.Buffer
		var filebuff bytes.Buffer
		archive := tar.NewWriter(&buff)

		for hash, submap := range db.hashmap {
			if err := gob.NewEncoder(&filebuff).Encode(submap.peers); err != nil {
				b.Error("Failed to encode a peermap", err, hash)
			}

			hdr := &tar.Header{
				Name: hex.EncodeToString(hash[:]),
				Size: int64(filebuff.Len()),
			}
			if err := archive.WriteHeader(hdr); err != nil {
				log.Fatal(err)
			}
			if _, err := archive.Write(filebuff.Bytes()); err != nil {
				b.Error(err)
			}

			filebuff.Reset()
		}

		if err := archive.Close(); err != nil {
			b.Error("Failed to close archive", err)
		}

		b.Log(len(buff.Bytes()))
	}
}

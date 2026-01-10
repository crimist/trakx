package main

import mrand "math/rand"

type dataset struct {
	torrents      [][]byte
	peers         [][]byte
	encodedHashes []string
	encodedPeers  []string
}

func newDataset(rng *mrand.Rand, torrents, peers int) *dataset {
	ds := &dataset{
		torrents:      make([][]byte, torrents),
		peers:         make([][]byte, peers),
		encodedHashes: make([]string, torrents),
		encodedPeers:  make([]string, peers),
	}
	for i := 0; i < torrents; i++ {
		buf := make([]byte, 20)
		rng.Read(buf)
		ds.torrents[i] = buf
		ds.encodedHashes[i] = percentEncode(buf)
	}
	for i := 0; i < peers; i++ {
		buf := make([]byte, 20)
		rng.Read(buf)
		ds.peers[i] = buf
		ds.encodedPeers[i] = percentEncode(buf)
	}
	return ds
}

func percentEncode(b []byte) string {
	const hex = "0123456789ABCDEF"
	out := make([]byte, len(b)*3)
	for i, v := range b {
		out[i*3] = '%'
		out[i*3+1] = hex[v>>4]
		out[i*3+2] = hex[v&0x0f]
	}
	return string(out)
}

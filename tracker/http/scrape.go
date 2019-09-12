package http

import (
	"net"
	"net/url"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/storage"
)

func (t *HTTPTracker) scrape(conn net.Conn, vals url.Values) {
	storage.AddExpval(&storage.Expvar.Scrapes, 1)

	infohashes := vals["info_hash"]
	if len(infohashes) == 0 {
		t.clientError(conn, "no infohashes")
		return
	}

	dict := bencoding.NewDict()
	nestedDict := make(map[string]map[string]int32, len(infohashes))

	for _, infohash := range infohashes {
		if len(infohash) != 20 {
			t.clientError(conn, "invalid infohash")
			return
		}

		nestedDict[infohash] = make(map[string]int32, 2)

		var hash storage.Hash
		copy(hash[:], infohash)

		complete, incomplete := t.peerdb.HashStats(&hash)

		nestedDict[infohash]["complete"] = complete
		nestedDict[infohash]["incomplete"] = incomplete
	}

	if err := dict.Any("files", nestedDict); err != nil {
		t.internalError(conn, "dict.Add()", err)
		return
	}

	conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + dict.Get()))
	storage.AddExpval(&storage.Expvar.ScrapesOK, 1)
}

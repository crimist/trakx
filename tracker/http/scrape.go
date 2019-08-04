package http

import (
	"sync/atomic"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
)

func (t *HTTPTracker) Scrape(c *ctx) {
	atomic.AddInt64(&shared.Expvar.Scrapes, 1)

	infohashes := c.u.Query()["info_hash"]
	if len(infohashes) == 0 {
		t.clientError(c.conn, "no infohashes")
		return
	}

	dict := bencoding.NewDict()
	nestedDict := make(map[string]map[string]int32)

	for _, infohash := range infohashes {
		if len(infohash) != 20 {
			t.clientError(c.conn, "invalid infohash")
			return
		}

		nestedDict[infohash] = make(map[string]int32)

		var hash shared.Hash
		copy(hash[:], infohash)

		complete, incomplete := t.peerdb.HashStats(&hash)

		nestedDict[infohash]["complete"] = complete
		nestedDict[infohash]["incomplete"] = incomplete
	}

	if err := dict.Add("files", nestedDict); err != nil {
		t.internalError(c.conn, "dict.Add()", err)
		return
	}

	c.conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + dict.Get()))
	atomic.AddInt64(&shared.Expvar.ScrapesOK, 1)
}

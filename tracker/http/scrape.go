package http

import (
	"net"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/storage"
)

func (t *HTTPTracker) scrape(conn net.Conn, infohashes params) {
	stats.Scrapes.Add(1)

	d := bencoding.GetDictionary()
	d.StartDictionary("files")

	for _, infohash := range infohashes {
		if infohash == nil {
			continue
		}
		if len(infohash) != 20 {
			t.clientError(conn, "invalid infohash")
			return
		}

		var hash storage.Hash
		copy(hash[:], infohash)
		complete, incomplete := t.peerdb.HashStats(hash)

		d.StartDictionaryBytes(infohash)
		{
			d.Int64("complete", int64(complete))
			d.Int64("incomplete", int64(incomplete))
		}
		d.EndDictionary()
	}

	d.EndDictionary()

	// conn.Write(httpSuccessBytes)
	// conn.Write(d.GetBytes())

	conn.Write(append(httpSuccessBytes, d.GetBytes()...))

	bencoding.PutDictionary(d)
}

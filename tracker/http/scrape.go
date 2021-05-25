package http

import (
	"net"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/storage"
)

func (t *HTTPTracker) scrape(conn net.Conn, infohashes params) {
	storage.Expvar.Scrapes.Add(1)
	defer storage.Expvar.ScrapesOK.Add(1)

	d := bencoding.NewDict()
	d.StartDict("files")

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

		d.StartDictBytes(infohash)
		{
			d.Int64("complete", int64(complete))
			d.Int64("incomplete", int64(incomplete))
		}
		d.EndDict()
	}

	d.EndDict()

	// conn.Write(httpSuccessBytes)
	// conn.Write(d.GetBytes())

	conn.Write(append(httpSuccessBytes, d.GetBytes()...))
}

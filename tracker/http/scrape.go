package http

import (
	"net"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
)

func (t *HTTPTracker) scrape(conn net.Conn, infohashes params) {
	storage.Expvar.Scrapes.Add(1)
	defer storage.Expvar.ScrapesOK.Add(1)

	root := bencoding.NewDict()

	for _, infohash := range infohashes {
		if infohash == "" {
			continue
		}
		if len(infohash) != 20 {
			t.clientError(conn, "invalid infohash")
			return
		}

		var hash storage.Hash
		copy(hash[:], infohash)
		complete, incomplete := t.peerdb.HashStats(hash)

		d := bencoding.NewDict()
		d.Int64("complete", int64(complete))
		d.Int64("incomplete", int64(incomplete))
		root.Dictionary(infohash, d.Get())
	}

	tmp := bencoding.NewDict()
	tmp.Dictionary("files", root.Get())

	conn.Write(shared.StringToBytes(httpSuccess + tmp.Get()))
}

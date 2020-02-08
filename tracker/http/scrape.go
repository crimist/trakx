package http

import (
	"net"
	"net/url"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
)

func (t *HTTPTracker) scrape(conn net.Conn, vals url.Values) {
	storage.AddExpval(&storage.Expvar.Scrapes, 1)

	infohashes := vals["info_hash"]
	if len(infohashes) == 0 {
		t.clientError(conn, "no infohashes")
		return
	}

	root := bencoding.NewDict()

	for _, infohash := range infohashes {
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
	storage.AddExpval(&storage.Expvar.ScrapesOK, 1)
}

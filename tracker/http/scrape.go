package http

import (
	"net"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/storage"
)

func (t *HTTPTracker) scrape(conn net.Conn, infohashes params) {
	stats.Scrapes.Add(1)

	dictionary := pools.Dictionaries.Get()
	dictionary.StartDictionary("files")

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

		dictionary.StartDictionaryBytes(infohash)
		{
			dictionary.Int64("complete", int64(complete))
			dictionary.Int64("incomplete", int64(incomplete))
		}
		dictionary.EndDictionary()
	}

	dictionary.EndDictionary()

	// conn.Write(httpSuccessBytes)
	// conn.Write(d.GetBytes())

	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))
	pools.Dictionaries.Put(dictionary)
}

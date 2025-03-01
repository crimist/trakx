package http

import (
	"net"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/storage"
)

func (tracker *Tracker) scrape(conn net.Conn, infohashes parameters) {
	tracker.stats.Scrapes.Add(1)

	dictionary := pools.Dictionaries.Get()
	dictionary.StartDictionary("files")

	for _, infohash := range infohashes {
		if infohash == nil {
			continue
		}
		if len(infohash) != 20 {
			writeFailure(conn, "invalid infohash")
			tracker.stats.ClientErrors.Add(1)
			return
		}

		var hash storage.Hash
		copy(hash[:], infohash)
		seeds, leeches := tracker.peerdb.TorrentStats(hash)

		dictionary.StartDictionaryBytes(infohash)
		{
			dictionary.Int64("complete", int64(seeds))
			dictionary.Int64("incomplete", int64(leeches))
		}
		dictionary.EndDictionary()
	}

	dictionary.EndDictionary()

	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))
	pools.Dictionaries.Put(dictionary)
}

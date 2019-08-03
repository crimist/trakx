package http

import (
	"net/http"
	"sync/atomic"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
)

func (t *HTTPTracker) ScrapeHandle(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&shared.Expvar.Scrapes, 1)

	infohashes := r.URL.Query()["info_hash"]
	if len(infohashes) == 0 {
		t.clientError("no infohashes", w)
		return
	}

	dict := bencoding.NewDict()
	nestedDict := make(map[string]map[string]int32)

	for _, infohash := range infohashes {
		if len(infohash) != 20 {
			t.clientError("invalid infohash", w)
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
		t.internalError("dict.Add()", err, w)
		return
	}

	w.Write([]byte(dict.Get()))
	atomic.AddInt64(&shared.Expvar.ScrapesOK, 1)
}

package http

import (
	"net/http"
	"sync/atomic"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
)

func ScrapeHandle(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&shared.ExpvarScrapes, 1)

	infohashes := r.URL.Query()["info_hash"]
	if len(infohashes) == 0 {
		clientError("no infohashes", w)
		return
	}

	dict := bencoding.NewDict()
	nestedDict := make(map[string]map[string]int32)

	for _, infohash := range infohashes {
		if len(infohash) != 20 {
			clientError("invalid infohash", w)
			return
		}

		nestedDict[infohash] = make(map[string]int32)

		var hash shared.Hash
		copy(hash[:], infohash)

		complete, incomplete := hash.Complete()

		nestedDict[infohash]["complete"] = complete
		nestedDict[infohash]["incomplete"] = incomplete
	}

	if err := dict.Add("files", nestedDict); err != nil {
		internalError("dict.Add()", err, w)
		return
	}

	w.Write([]byte(dict.Get()))
	atomic.AddInt64(&shared.ExpvarScrapesOK, 1)
}

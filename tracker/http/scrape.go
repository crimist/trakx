package http

import (
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/tracker/shared"
	"github.com/go-torrent/bencode"
)

func ScrapeHandle(w http.ResponseWriter, r *http.Request) {
	shared.ExpvarScrapes++

	infohashes := r.URL.Query()["info_hash"]
	if len(infohashes) == 0 {
		clientError("no infohashes", w)
		return
	}

	dict := bencode.Dictionary{
		"files": bencode.Dictionary{},
	}

	for _, infohash := range infohashes {
		if len(infohash) != 20 {
			clientError("invalid infohash", w)
			return
		}

		var hash shared.Hash
		copy(hash[:], infohash)

		complete, incomplete := hash.Complete()

		nested := dict["files"].(bencode.Dictionary)
		nested[infohash] = bencode.Dictionary{
			"complete":   complete,
			"incomplete": incomplete,
		}
	}

	resp, err := bencode.Marshal(dict)
	if err != nil {
		internalError("bencode.Marshal", err, w)
		return
	}

	fmt.Fprint(w, string(resp))
}

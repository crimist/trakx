package http

import (
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/tracker"
	"github.com/go-torrent/bencode"
	"go.uber.org/zap"
)

func clientError(writer http.ResponseWriter, reason string) {
	dict := bencode.Dictionary{
		"failure reason": reason,
	}
	data, _ := bencode.Marshal(dict)
	fmt.Fprint(writer, string(data))
}

func ScrapeHandle(w http.ResponseWriter, r *http.Request) {
	tracker.ExpvarScrapes++

	infohashes := r.URL.Query()["info_hash"]
	if len(infohashes) == 0 {
		clientError(w, "no infohashes")
		return
	}

	dict := bencode.Dictionary{
		"files": bencode.Dictionary{},
	}

	for _, infohash := range infohashes {
		if len(infohash) != 20 {
			clientError(w, "invalid infohash")
			return
		}

		var hash tracker.Hash
		var complete, incomplete int
		copy(hash[:], infohash)

		for _, peer := range tracker.PeerDB[hash] {
			if peer.Complete {
				complete++
			} else {
				incomplete++
			}
		}

		nested := dict["files"].(bencode.Dictionary)
		nested[infohash] = bencode.Dictionary{
			"complete":   complete,
			"incomplete": incomplete,
		}
	}

	resp, err := bencode.Marshal(dict)
	if err != nil {
		tracker.Logger.Error("bencode", zap.Error(err))
	}

	fmt.Fprint(w, string(resp))
}

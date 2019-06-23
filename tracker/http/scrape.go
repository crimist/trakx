package http

import (
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/tracker/shared"
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
	shared.ExpvarScrapes++

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
		shared.Logger.Error("bencode", zap.Error(err))
	}

	fmt.Fprint(w, string(resp))
}

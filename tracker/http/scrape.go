package http

import (
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/tracker/shared"
	"github.com/Syc0x00/Trakx/bencoding"
)

func ScrapeHandle(w http.ResponseWriter, r *http.Request) {
	shared.ExpvarScrapes++

	infohashes := r.URL.Query()["info_hash"]
	if len(infohashes) == 0 {
		clientError("no infohashes", w)
		return
	}

	d := bencoding.NewDict()
	files := make(map[string]int32)

	for _, infohash := range infohashes {
		if len(infohash) != 20 {
			clientError("invalid infohash", w)
			return
		}

		var hash shared.Hash
		copy(hash[:], infohash)

		complete, incomplete := hash.Complete()

		files["complete"] = complete
		files["incomplete"] = incomplete
	}

	if err := d.Add("files", files); err != nil {
		internalError("dict.Add()", err, w)
		return
	}

	fmt.Fprint(w, d.Get())
}

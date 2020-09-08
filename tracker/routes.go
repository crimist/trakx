package tracker

import (
	"net/http"

	"github.com/crimist/trakx/tracker/shared"
)

func index(w http.ResponseWriter, r *http.Request) {
	w.Write(shared.IndexDataBytes)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	w.Write(shared.DMCADataBytes)
}

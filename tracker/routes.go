package tracker

import (
	"net/http"

	"github.com/crimist/trakx/tracker/config"
)

func index(w http.ResponseWriter, r *http.Request) {
	w.Write(config.IndexDataBytes)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	w.Write(config.DMCADataBytes)
}

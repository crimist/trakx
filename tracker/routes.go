package tracker

import (
	"io/ioutil"
	"net/http"

	"github.com/Syc0x00/Trakx/tracker/shared"
)

var indexData []byte

func initRoutes() {
	var err error
	if indexData, err = ioutil.ReadFile(shared.Config.Tracker.Index); err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexData)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=BwSts2s4ba4", http.StatusTemporaryRedirect)
}

func stats(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(shared.StatsHTML))
}

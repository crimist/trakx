package tracker

import (
	"io/ioutil"
	"net/http"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

var indexData []byte

func initRoutes() {
	var err error
	if indexData, err = ioutil.ReadFile(conf.Trakx.Index); err != nil {
		logger.Panic("Failed to read index", zap.Error(err))
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

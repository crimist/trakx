package tracker

import (
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/syc0x00/trakx/tracker/shared"
)

var indexData []byte

func initRoutes() {
	var err error
	if indexData, err = ioutil.ReadFile(shared.Config.Trakx.Index); err != nil {
		shared.Logger.Panic("Failed to read index", zap.Error(err))
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

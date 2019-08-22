package tracker

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	"go.uber.org/zap"
)

var indexData []byte

func initRoutes() {
	var err error
	if indexData, err = ioutil.ReadFile(conf.Trakx.Index); err != nil {
		logger.Info("Failed to read index", zap.Error(err))

		home, err := os.UserHomeDir()
		if err != nil {
			logger.Panic("Failed to get home", zap.Error(err))
		}
		if indexData, err = ioutil.ReadFile(home + "/index.html"); err != nil {
			logger.Panic("Failed to read index", zap.Error(err))
		}
	}
	indexData = bytes.ReplaceAll(indexData, []byte("\t"), nil)
	indexData = bytes.ReplaceAll(indexData, []byte("\n"), nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexData)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=BwSts2s4ba4", http.StatusTemporaryRedirect)
}

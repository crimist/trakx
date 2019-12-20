package tracker

import (
	"bytes"
	"io/ioutil"
	"net/http"

	trakxhttp "github.com/crimist/trakx/tracker/http"
	"go.uber.org/zap"
)

var indexData []byte

func initRoutes() {
	var err error
	if indexData, err = ioutil.ReadFile(conf.Trakx.Index); err != nil {
		logger.Panic("Failed to read index", zap.Error(err))
	}
	indexData = bytes.ReplaceAll(indexData, []byte("\t"), nil)
	indexData = bytes.ReplaceAll(indexData, []byte("\n"), nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexData)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(trakxhttp.DMCAData))
}

package tracker

import (
	"net/http"
	"io/ioutil"
)

var indexData []byte

func initRoutes() {
	var err error
	if indexData, err = ioutil.ReadFile("tracker/index.html"); err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexData)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=BwSts2s4ba4", http.StatusMovedPermanently)
}
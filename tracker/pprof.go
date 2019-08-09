package tracker

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func initpprof() {
	// only listen on localhost
	go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", conf.Trakx.Pprof.Port), nil)
}

package tracker

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/crimist/trakx/tracker/config"
)

func runpprof() {
	// only listen on localhost
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", config.Conf.Debug.PprofPort), nil)
}

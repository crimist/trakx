package tracker

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/crimist/trakx/tracker/config"
	"go.uber.org/zap"
)

func servePprof() {
	config.Logger.Info("Serving pprof", zap.Int("port", config.Config.Debug.Pprof))

	// serve on localhost
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", config.Config.Debug.Pprof), nil)
}

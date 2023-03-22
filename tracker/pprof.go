package tracker

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"go.uber.org/zap"
)

func servePprof(port int) {
	zap.L().Info("Serving pprof", zap.Int("port", port))

	// serve on localhost
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
}

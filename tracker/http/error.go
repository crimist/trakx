package http

import (
	"net/http"
	"sync/atomic"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

func writeErr(msg string, writer http.ResponseWriter) {
	d := bencoding.NewDict()
	d.Add("failure reason", msg)
	writer.Write([]byte(d.Get()))
}

func (t *HTTPTracker) clientError(msg string, writer http.ResponseWriter, fields ...zap.Field) {
	if !t.conf.Trakx.Prod {
		fields = append(fields, zap.String("msg", msg))
		t.logger.Info("Client Error", fields...)
	}

	atomic.AddInt64(&shared.Expvar.Clienterrs, 1)
	writeErr(msg, writer)
}

func (t *HTTPTracker) internalError(errmsg string, err error, writer http.ResponseWriter) {
	atomic.AddInt64(&shared.Expvar.Errs, 1)
	writeErr("internal server error", writer)
	t.logger.Error(errmsg, zap.Error(err))
}

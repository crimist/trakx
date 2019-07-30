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

func clientError(msg string, writer http.ResponseWriter, fields ...zap.Field) {
	atomic.AddInt64(&shared.ExpvarClienterrs, 1)
	writeErr(msg, writer)
	if !shared.Config.Trakx.Prod {
		fields = append(fields, zap.String("msg", msg))
		shared.Logger.Info("Client Error", fields...)
	}
}

func internalError(errmsg string, err error, writer http.ResponseWriter) {
	atomic.AddInt64(&shared.ExpvarErrs, 1)
	writeErr("internal server error", writer)
	shared.Logger.Error(errmsg, zap.Error(err))
}

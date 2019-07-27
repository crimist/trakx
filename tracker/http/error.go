package http

import (
	"net/http"

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
	shared.ExpvarClienterrs++
	writeErr(msg, writer)
	if shared.Env == shared.Dev {
		fields = append(fields, zap.String("msg", msg))
		shared.Logger.Info("Client Error", fields...)
	}
}

func internalError(errmsg string, err error, writer http.ResponseWriter) {
	shared.ExpvarErrs++
	writeErr("internal server error", writer)
	shared.Logger.Error(errmsg, zap.Error(err))
}

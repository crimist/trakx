package http

import (
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/tracker/shared"
	"go.uber.org/zap"
)

func writeErr(msg string, writer http.ResponseWriter) {
	d := bencoding.NewDict()
	d.Add("failure reason", msg)
	fmt.Fprint(writer, d.Get())
}

func clientError(msg string, writer http.ResponseWriter, fields ...zap.Field) {
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

//go:build heroku
// +build heroku

package http

import (
	"bytes"
)

func parseForwarded(data []byte) (bool, []byte) {
	headerKey := []byte("X-Forwarded-For: ")

	headerValueStart := bytes.Index(data, headerKey)
	if headerValueStart == -1 {
		return false, nil
	}
	headerValueStart += len(headerKey)

	headerValueEnd := bytes.Index(data[headerValueStart:], []byte("\r\n"))
	if headerValueEnd == -1 {
		return true, nil
	}
	headerValueEnd += headerValueStart

	headerValue := data[headerValueStart:headerValueEnd]

	// if list contains multiple addresses use rightmost (most recent)
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	if bytes.Contains(headerValue, []byte(",")) {
		lastItemIndex := bytes.LastIndex(headerValue, []byte(",")) + 2
		return true, headerValue[lastItemIndex:]
	}

	return true, headerValue
}

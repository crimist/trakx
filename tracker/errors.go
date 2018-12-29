package tracker

import (
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/bencoding"
)

// TrackErr if for tracker errors
type TrackErr uint8

const (
	// OK means success
	OK TrackErr = 0
	// Error is a generic error
	Error TrackErr = 1
	// Banned means the hash is banned
	Banned TrackErr = 2
)

// ThrowErr throws a tracker bencoded error to the client
func ThrowErr(w http.ResponseWriter, reason string, status int) {
	d := bencoding.NewDict()
	d.Add("failure reason", reason)
	w.WriteHeader(status)
	fmt.Fprint(w, d.Get())
}

// InternalError is a wrapper to tell the client I fucked up
func InternalError(w http.ResponseWriter) {
	ThrowErr(w, "Internal Server Error", http.StatusInternalServerError)
}

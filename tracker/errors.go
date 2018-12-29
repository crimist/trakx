package tracker

// TrackErr if for tracker errors
type TrackErr uint8

const (
	// OK means success
	OK TrackErr = iota
	// Error is a generic error
	Error
	// Banned means the hash is banned
	Banned
)

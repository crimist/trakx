package config

// LogLevel holds designated logging level
type LogLevel string

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

// Debug returns true if the loglevel is set to debug.
func (l LogLevel) Debug() bool {
	return l == "debug"
}

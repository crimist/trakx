package config

import (
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type Configuration struct {
	LogLevel       LogLevel
	CachePath      string
	ExpvarInterval time.Duration
	Debug          struct {
		Pprof int
	}
	Announce struct {
		Base time.Duration
		Fuzz time.Duration
	}
	HTTP struct {
		Mode    string
		IP      string
		Port    int
		Timeout struct {
			Read  time.Duration
			Write time.Duration
		}
		Threads int
	}
	UDP struct {
		Enabled bool
		IP      string
		Port    int
		Threads int
		ConnDB  struct {
			Validate bool
			Size     uint64
			Trim     time.Duration
			Expiry   time.Duration
		}
	}
	Numwant struct {
		Default uint
		Limit   uint
	}
	DB struct {
		Type   string
		Backup struct {
			Frequency time.Duration
			Type      string
			Path      string
		}
		Trim   time.Duration
		Expiry time.Duration
	}
}

// LogPath returns the log path as defined by the configuration and current time
func (conf *Configuration) LogPath() string {
	return filepath.Join(conf.CachePath, "trakx_"+time.Now().Format("06-01-02-15-04-05")+".log")
}

// PIDPath retuirns the pid file path
func (conf *Configuration) PIDPath() string {
	return filepath.Join(conf.CachePath, "trakx.pid")
}

// setLogLevel sets the desired loglevel in the in memory configuration and logger
func (conf *Configuration) setLogLevel(level LogLevel) {
	conf.LogLevel = level

	switch level {
	case "debug":
		loggerAtom.SetLevel(zap.DebugLevel)
		zap.L().Debug("Debug loglevel set, debug panics enabled")
	case "info":
		loggerAtom.SetLevel(zap.InfoLevel)
	case "warn":
		loggerAtom.SetLevel(zap.WarnLevel)
	case "error":
		loggerAtom.SetLevel(zap.ErrorLevel)
	case "fatal":
		loggerAtom.SetLevel(zap.FatalLevel)
	default:
		zap.L().Warn("Invalid log level was specified, defaulting to warn")
		loggerAtom.SetLevel(zap.WarnLevel)
	}
}

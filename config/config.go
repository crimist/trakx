package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type LogLevel string

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

type Configuration struct {
	LogLevel LogLevel
	Cache    string
	Stats    struct {
		General  bool
		IP       bool
		Interval time.Duration
	}
	Debug struct {
		Pprof int
	}
	Announce struct {
		Base time.Duration
		Fuzz time.Duration
	}
	HTTP struct {
		Tracker  bool
		Port     int
		IP       string
		Routines int
		Timeout  struct {
			Read  time.Duration
			Write time.Duration
		}
		Serve string
	}
	UDP struct {
		Port        int
		IP          string
		Routines    int
		Connections struct {
			Validate bool
			GC       time.Duration
			Expiry   time.Duration
		}
	}
	Numwant struct {
		Default uint
		Limit   uint
	}
	DB struct {
		GC     time.Duration
		Expiry time.Duration
		Backup struct {
			Interval time.Duration
			Path     string
		}
	}
}

// validate checks that configuration values are sane and returns warnings for potential misconfigurations or security issues
func (conf *Configuration) validate() error {
	// check cache directory exists and is a directory
	stat, err := os.Stat(conf.Cache)
	if os.IsNotExist(err) {
		zap.L().Debug("cache directory does not exist, creating it", zap.String("cache", conf.Cache))
		if err = os.MkdirAll(conf.Cache, defaultFolderPermission); err != nil {
			return errors.Wrapf(err, "failed to create cache directory '%s'", conf.Cache)
		}
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat cache '%s'", conf.Cache)
	} else if !stat.IsDir() {
		return errors.Errorf("cache '%s' is not a directory", conf.Cache)
	}

	if !conf.UDP.Connections.Validate {
		zap.L().Warn("Configuration warning: UDP connection validation is disabled. Do not expose this service to untrusted networks; it could be abused in UDP based amplification attacks.")
	}

	if conf.DB.Expiry < conf.Announce.Base+conf.Announce.Fuzz {
		zap.L().Warn("Configuration warning: Peer expiry time < announce base + fuzz. Peers will expire from the database between announces.")
	}

	if conf.Stats.General {
		if conf.Stats.Interval <= 0 {
			zap.L().Fatal("Invalid configuration: Stats.Interval must be greater than 0 if Stats.General is enabled")
		}

		if conf.HTTP.Port == 0 {
			zap.L().Warn("Configuration warning: Statistics collection enabled but no HTTP server is running to publish them")
		}
	}

	return nil
}

// LogPath returns the log path as defined by the configuration and current time
func (conf *Configuration) LogPath() string {
	return filepath.Join(conf.Cache, "trakx_"+time.Now().Format("06-01-02-15-04-05")+".log")
}

// PIDPath retuirns the pid file path
func (conf *Configuration) PIDPath() string {
	return filepath.Join(conf.Cache, "trakx.pid")
}

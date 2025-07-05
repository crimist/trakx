package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
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
		Mode    string
		IP      string
		Port    int
		Timeout struct {
			Read  time.Duration
			Write time.Duration
		}
		Threads   int
		ServePath string
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

	// potential misconfigurations warnings
	if !conf.UDP.ConnDB.Validate {
		zap.L().Warn("Configuration warning: UDP connection validation is disabled. Do not expose this service to untrusted networks; it could be abused in UDP based amplification attacks.")
	}
	if conf.DB.Expiry < conf.Announce.Base+conf.Announce.Fuzz {
		zap.L().Warn("Configuration warning: Peer expiry time < announce base + fuzz. Peers will expire from the database between announces.")
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

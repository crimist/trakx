package config

import (
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	LogLevel    LogLevel
	Pprof       int
	Expvar      time.Duration
	NofileLimit uint64
	Announce    struct {
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
		Size   uint64
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

// Loaded returns true if the config was successfully loaded.
func (conf *Config) Loaded() bool {
	// Database.Type is required to run so if it's empty we know that the config isn't loaded
	return conf.Database.Type != ""
}

// SetLogLevel sets the desired loglevel in the in memory configuration and logger
func (conf *Config) SetLogLevel(level LogLevel) {
	conf.LogLevel = level

	switch level {
	case "debug":
		loggerAtom.SetLevel(zap.DebugLevel)
		Logger.Debug("Debug loglevel set, debug panics enabled")
	case "info":
		loggerAtom.SetLevel(zap.InfoLevel)
	case "warn":
		loggerAtom.SetLevel(zap.WarnLevel)
	case "error":
		loggerAtom.SetLevel(zap.ErrorLevel)
	case "fatal":
		loggerAtom.SetLevel(zap.FatalLevel)
	default:
		Logger.Warn("Invalid log level was specified, defaulting to warn")
		loggerAtom.SetLevel(zap.WarnLevel)
	}
}

var oneTimeSetup sync.Once

// Update updates logger and limits based on the configuration settings.
func (conf *Config) Update() error {
	oneTimeSetup.Do(func() {
		loggerAtom = zap.NewAtomicLevelAt(zap.DebugLevel)
	})

	cfg := zap.NewDevelopmentConfig()

	// set LogLevel to lower case (casting nightmare)
	conf.LogLevel = LogLevel(strings.ToLower(string(conf.LogLevel)))
	conf.Tracker.HTTP.Mode = strings.ToLower(conf.Tracker.HTTP.Mode)

	if conf.LogLevel.Debug() {
		cfg.Development = true
	} else {
		cfg.Development = false
	}

	Logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.Lock(os.Stdout), loggerAtom))
	conf.SetLogLevel(conf.LogLevel)

	// limits
	if conf.Debug.NofileLimit != nofileIgnore {
		var rLimit syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			return errors.Wrap(err, "failed to get the NOFILE limit")
		}
		Logger.Debug("Got nofile limits", zap.Any("limit", rLimit))

		// Limit is bugged on WSL and Darwin systems, to avoid bug keep limit below 10_000
		if ulimitBugged() && conf.Debug.NofileLimit > 10000 {
			Logger.Warn("Detected bugged nofile limit, you are on Darwin or WSL based systen. Capping nofile limit to 10_000.")
			rLimit.Max = 10000
			rLimit.Cur = 10000
		} else {
			rLimit.Max = conf.Debug.NofileLimit
			rLimit.Cur = conf.Debug.NofileLimit
		}

		Logger.Debug("Setting nofile limit", zap.Any("limit", rLimit))
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			return errors.Wrap(err, "failed to set the NOFILE limit")
		}
	}

	return nil
}

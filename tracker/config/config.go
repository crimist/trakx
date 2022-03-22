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
	LogLevel LogLevel
	Debug    struct {
		PprofPort      int
		ExpvarInterval time.Duration
		NofileLimit    uint64
		PeerChanInit   uint64
		CheckConnIDs   bool
	}
	Tracker struct {
		Announce     time.Duration
		AnnounceFuzz time.Duration
		HTTP         struct {
			Mode         string
			Port         int
			ReadTimeout  time.Duration
			WriteTimeout time.Duration
			Threads      int
		}
		UDP struct {
			Enabled bool
			Port    int
			Threads int
		}
		Numwant struct {
			Default int32
			Limit   int32
		}
	}
	Database struct {
		Type    string
		Backup  string
		Address string
		Peer    struct {
			Trim    time.Duration
			Write   time.Duration
			Timeout time.Duration
		}
		Conn struct {
			Trim    time.Duration
			Timeout time.Duration
		}
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
		Logger.Debug("Debug level enabled, debug panics are on")
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

	Logger.Debug("Set log level", zap.Any("level", level))
}

var oneTimeSetup sync.Once

// Update updates logger and ulimited based on config.
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
	Logger.Debug("logger created")

	conf.SetLogLevel(conf.LogLevel)

	// limits
	if conf.Debug.NofileLimit != nofileIgnore {
		var rLimit syscall.Rlimit
		err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Wrap(err, "failed to get the NOFILE limit")
		}

		Logger.Debug("Got nofile limit", zap.Any("limit", rLimit))

		// Bugged on OSX & WSL
		if ulimitBugged() && conf.Debug.NofileLimit > 10000 {
			Logger.Debug("Detected bugged rlimit, capping to 10'000")
			rLimit.Max = 10000
			rLimit.Cur = 10000
		} else {
			rLimit.Max = conf.Debug.NofileLimit
			rLimit.Cur = conf.Debug.NofileLimit
		}

		Logger.Debug("Setting nofile limit", zap.Any("limit", rLimit))

		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Wrap(err, "failed to set the NOFILE limit")
		}
	}

	return nil
}

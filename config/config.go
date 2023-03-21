package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Configuration struct {
	loaded bool // config is loaded and valid

	LogLevel       LogLevel
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
	Path struct {
		Log string
		Pid string
	}
}

// Loaded returns true if the config was successfully parsed and loaded.
func (config *Configuration) Loaded() bool { return config.loaded }

// SetLogLevel sets the desired loglevel in the in memory configuration and logger
func (conf *Configuration) SetLogLevel(level LogLevel) {
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

// Parse updates logger and limits based on the configuration settings.
func (config *Configuration) Parse() error {
	// one time logger atom setup
	oneTimeSetup.Do(func() {
		loggerAtom = zap.NewAtomicLevelAt(zap.DebugLevel)
	})

	cfg := zap.NewDevelopmentConfig()

	// set strings to lowercase
	config.LogLevel = LogLevel(strings.ToLower(string(config.LogLevel)))
	config.HTTP.Mode = strings.ToLower(config.HTTP.Mode)

	// dev env check
	if config.LogLevel.Debug() {
		cfg.Development = true
	} else {
		cfg.Development = false
	}

	// setup logger
	Logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.Lock(os.Stdout), loggerAtom))
	config.SetLogLevel(config.LogLevel)

	// resolve env vars for database backup path
	if strings.HasPrefix(config.DB.Backup.Path, "ENV:") {
		config.DB.Backup.Path = os.Getenv(strings.TrimPrefix(config.DB.Backup.Path, "ENV:"))
	}

	// resolve paths
	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Fatal("failed to get home directory", zap.Error(err))
	}
	config.Path.Pid = strings.ReplaceAll(config.Path.Pid, "~", home)
	config.Path.Log = strings.ReplaceAll(config.Path.Log, "~", home)

	// If $PORT var set override port for appengines (like heroku)
	if appenginePort := os.Getenv("PORT"); appenginePort != "" {
		appPort, err := strconv.Atoi(appenginePort)
		if err != nil {
			return errors.Wrap(err, "failed to parse $PORT env variable (not an int)")
		}

		Logger.Info("PORT env variable detected. Overriding config...", zap.Int("$PORT", appPort))
		config.HTTP.Port = appPort
	}

	config.loaded = true
	return nil
}

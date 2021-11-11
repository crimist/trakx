/*
	Config holds configuration information for trakx.
*/
package config

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/kkyr/fig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	nofileIgnore        = 0
	TrackerModeEnabled  = "enabled"  // http tracker enabled
	TrackerModeInfo     = "info"     // http information server, no tracker
	TrackerModeDisabled = "disabled" // http disabled
)

var (
	// Global instance of config and logger
	Conf   *Config
	Logger *zap.Logger

	loggerAtom zap.AtomicLevel
)

func init() {
	// create temporary logger
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic("failed to init initial zap logger")
	}

	// load paths
	initPaths()

	// load embedded filesystem
	loadEmbed()

	// load config
	Conf, err = Load()
	if err != nil {
		Logger.Error("Failed to load a config", zap.Any("config", Conf), zap.Error(err))
	} else {
		if Conf.LogLevel.Debug() {
			Logger.Debug("Loaded config", zap.Any("config", Conf))
		} else {
			Logger.Info("Loaded config")
		}
	}

	Logger.Debug("initialized paths", zap.String("config", ConfigDir), zap.String("cache", CacheDir))
}

// LogLevel holds designated logging level
type LogLevel string

// Debug returns true if the loglevel is set to debug.
func (l LogLevel) Debug() (dbg bool) {
	if l == "debug" {
		dbg = true
	}

	return
}

type Config struct {
	LogLevel LogLevel
	Debug    struct {
		PprofPort      int
		ExpvarInterval int
		NofileLimit    uint64
		PeerChanInit   uint64
		CheckConnIDs   bool
	}
	Tracker struct {
		Announce     int32
		AnnounceFuzz int32
		HTTP         struct {
			Mode         string
			Port         int
			ReadTimeout  int
			WriteTimeout int
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
			Trim    int
			Write   int
			Timeout int64
		}
		Conn struct {
			Trim    int
			Timeout int64
		}
	}
}

// Loaded returns true if the config was successfully loaded.
func (conf *Config) Loaded() bool {
	// Database.Type is required to run so if it's empty we know that the config isn't loaded
	return conf.Database.Type != ""
}

// Update updates logger and ulimited based on config.
func (conf *Config) Update() error {
	// logger and loglvl
	loggerAtom = zap.NewAtomicLevel()
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

	switch conf.LogLevel {
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

	Logger.Debug("logger created", zap.Any("loglevel", conf.LogLevel))

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

// Load attempts to load the config from the disk or environment.
// The config file must be named "trakx.yaml".
// Load searches for the config file in ".", "~/.config/trakx", "./embedded", "/app/embedded" in order.
// Environment variables overwrite file configuration, see trakx.yaml in ./embedded for examples.
// This function is called when the config package is imported.
func Load() (*Config, error) {
	conf := new(Config)

	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Error("Failed to get user home dir, attempting to continue config load", zap.Error(err))
	}

	err = fig.Load(conf,
		fig.File("trakx.yaml"),
		fig.UseEnv("trakx"),
		fig.Dirs(".", home+"/.config/trakx", "./embedded", "/app/embedded"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "fig failed to load a config")
	}

	// If $PORT var set override port for appengines (like heroku)
	if appenginePort := os.Getenv("PORT"); appenginePort != "" {
		appPort, err := strconv.Atoi(appenginePort)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse $PORT env variable (not an int)")
		}

		Logger.Info("PORT env variable detected. Overriding config...", zap.Int("$PORT", appPort))
		conf.Tracker.HTTP.Port = appPort
	}

	return conf, conf.Update()
}

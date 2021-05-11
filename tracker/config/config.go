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
	nofileIgnore = 0
)

var (
	// global instance of the config and logger
	Conf   *Config
	Logger *zap.Logger

	loggerAtom zap.AtomicLevel
)

func init() {
	// create temporary logger
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic("failed to init inital zap logger")
	}

	// load embeded filesystem
	loadEmbed()

	// load config
	Conf, err = Load()
	if err != nil {
		Logger.Error("Failed to load a config", zap.Any("config", Conf), zap.Error(err))
	} else {
		Logger.Info("Loaded config")
	}
}

// LogLevel is the logging level
type LogLevel string

// Debug checks whether the loglevel is set to debug
func (l LogLevel) Debug() (dbg bool) {
	if l == "debug" {
		dbg = true
	}
	return
}

type Config struct {
	LogLevel       LogLevel
	ExpvarInterval int
	PprofPort      int
	Ulimit         uint64
	PeerChanMin    uint64
	Tracker        struct {
		Announce     int32
		AnnounceFuzz int32
		HTTP         struct {
			Enabled      bool
			Port         int
			ReadTimeout  int
			WriteTimeout int
			Threads      int
		}
		UDP struct {
			Enabled     bool
			Port        int
			CheckConnID bool
			Threads     int
		}
		Numwant struct {
			Default int32
			Limit   int32
		}
	}
	Database struct {
		Type   string
		Backup string
		Peer   struct {
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

// Loaded returns true if the config was successfully loaded
func (conf *Config) Loaded() bool {
	// Database.Type is required to run so if it's empty we know that the config isn't loaded
	return conf.Database.Type != ""
}

// Update creates a logger with the given `config.LogLevel` and sets the desired ulimit
func (conf *Config) Update() error {
	// logger and loglvl
	loggerAtom = zap.NewAtomicLevel()
	cfg := zap.NewDevelopmentConfig()

	if conf.LogLevel.Debug() {
		cfg.Development = true
	} else {
		cfg.Development = false
	}

	Logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.Lock(os.Stdout), loggerAtom))

	switch strings.ToLower(string(conf.LogLevel)) {
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
	if conf.Ulimit != nofileIgnore {
		var rLimit syscall.Rlimit
		err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Wrap(err, "failed to get the NOFILE limit")
		}

		// Bugged on OSX & WSL
		if ulimitBugged() && conf.Ulimit > 10000 {
			rLimit.Max = 10000
			rLimit.Cur = 10000
		} else {
			rLimit.Max = conf.Ulimit
			rLimit.Cur = conf.Ulimit
		}
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Wrap(err, "failed to set the NOFILE limit")
		}
	}

	return nil
}

// Load attempts to load the config from the disk or environment
// It is called automatically when this package is imported
func Load() (*Config, error) {
	conf := new(Config)

	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Error("Failed to get user home dir, attempting to continue config load", zap.Error(err))
	}

	err = fig.Load(conf,
		fig.File("trakx.yaml"),
		fig.UseEnv("trakx"),
		fig.Dirs(".", home+"/.config/trakx", "./install", "/app/install"),
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

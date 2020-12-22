package shared

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
	// internal
	LoggerAtom zap.AtomicLevel
	Logger     *zap.Logger

	// configurable
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

func (conf *Config) update() error {
	// logger and loglvl
	if conf.Logger == nil {
		conf.LoggerAtom = zap.NewAtomicLevel()
		cfg := zap.NewDevelopmentConfig()

		if conf.LogLevel.Debug() {
			cfg.Development = true
		} else {
			cfg.Development = false
		}

		conf.Logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.Lock(os.Stdout), conf.LoggerAtom))
	}

	switch strings.ToLower(string(conf.LogLevel)) {
	case "debug":
		conf.LoggerAtom.SetLevel(zap.DebugLevel)
		conf.Logger.Debug("Debug level enabled, debug panics are on")
	case "info":
		conf.LoggerAtom.SetLevel(zap.InfoLevel)
	case "warn":
		conf.LoggerAtom.SetLevel(zap.WarnLevel)
	case "error":
		conf.LoggerAtom.SetLevel(zap.ErrorLevel)
	case "fatal":
		conf.LoggerAtom.SetLevel(zap.FatalLevel)
	default:
		conf.Logger.Warn("Invalid log level was specified, defaulting to warn")
		conf.LoggerAtom.SetLevel(zap.WarnLevel)
	}

	conf.Logger.Debug("logger created", zap.Any("loglevel", conf.LogLevel))

	// limits
	if conf.Ulimit == nofileIgnore {
		return nil
	}

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

	return nil
}

// LoadConf attempts to load the config from the disk or environment
func LoadConf() (*Config, error) {
	conf := new(Config)

	home, err := os.UserHomeDir()
	if err != nil {
		println("[ERROR] Failed to get user home dir, attempting to continue config load ->", err)
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

		conf.Tracker.HTTP.Port = appPort
	}

	return conf, conf.update()
}

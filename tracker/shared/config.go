package shared

import (
	"os"
	"strconv"
	"syscall"

	"github.com/kkyr/fig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	nofileIgnore = 0
)

type Config struct {
	Logger *zap.Logger
	Trakx  struct {
		Prod   bool
		Expvar struct {
			Every int
		}
		Pprof struct {
			Port int
		}
		Ulimit uint64
	}
	Tracker struct {
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
	// limits
	if conf.Trakx.Ulimit == nofileIgnore {
		return nil
	}

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return errors.Wrap(err, "failed to get the NOFILE limit")
	}

	// Bugged on OSX & WSL
	if ulimitBugged() && conf.Trakx.Ulimit > 10000 {
		rLimit.Max = 10000
		rLimit.Cur = 10000
	} else {
		rLimit.Max = conf.Trakx.Ulimit
		rLimit.Cur = conf.Trakx.Ulimit
	}
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return errors.Wrap(err, "failed to set the NOFILE limit")
	}

	return nil
}

// LoadConf attempts to load the config from the disk or environment
func LoadConf(logger *zap.Logger) (*Config, error) {
	conf := new(Config)
	conf.Logger = logger

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("failed to get user home dir", zap.Error(err))
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
			return nil, errors.Wrap(err, "failed to parse $PORT env variable as int")
		}

		conf.Tracker.HTTP.Port = appPort
	}

	return conf, conf.update()
}

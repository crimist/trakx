package shared

import (
	"os"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	TrakxRoot    = "/usr/local/etc/trakx/"
	nofileIgnore = 0
)

type Config struct {
	Logger *zap.Logger
	Trakx  struct {
		Prod   bool
		Index  string
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
			Qsize        int
			Accepters    int
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
			Filename string
			Trim     int
			Write    int
			Timeout  int64
		}
		Conn struct {
			Filename string
			Trim     int
			Timeout  int64
		}
	}
}

// Loaded checks if the config is loaded or not
func (conf *Config) Loaded() bool {
	// Database.Type is required to run so if it's empty we know that the config isn't loaded
	return conf.Database.Type != ""
}

func (conf *Config) update() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "failed to find home directory")
	}

	// update tild (~) with the actual home directory
	conf.Database.Conn.Filename = strings.ReplaceAll(conf.Database.Conn.Filename, "~", home)
	conf.Database.Peer.Filename = strings.ReplaceAll(conf.Database.Peer.Filename, "~", home)
	conf.Trakx.Index = strings.ReplaceAll(conf.Trakx.Index, "~", home)

	// limits
	if conf.Trakx.Ulimit == nofileIgnore {
		return nil
	}

	var rLimit syscall.Rlimit
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
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

func LoadConf(logger *zap.Logger) (*Config, error) {
	conf := new(Config)
	conf.Logger = logger

	// Load from file
	viper.SetConfigType("yaml")
	viper.SetConfigName("trakx")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app/")
	viper.AddConfigPath("/usr/local/etc/trakx/")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "viper failed to read config from disk")
	}

	// Load all env vars and override file - https://github.com/spf13/viper/issues/188#issuecomment-413368673
	viper.AutomaticEnv()
	viper.SetEnvPrefix("trakx")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.Unmarshal(conf); err != nil {
		return nil, errors.Wrap(err, "viper failed to unmarshal")
	}

	// If $PORT var set override everything for appengines (heroku)
	viper.BindEnv("app_port", "PORT")
	if appenginePort := viper.GetInt("app_port"); appenginePort != 0 {
		conf.Tracker.HTTP.Port = appenginePort
	}

	// Add watcher
	viper.OnConfigChange(func(e fsnotify.Event) {
		conf.Logger.Info("Config changed", zap.String("name", e.Name), zap.Any("op", e.Op))

		if err := viper.Unmarshal(conf); err != nil {
			conf.Logger.Info("Viper failed to unmarshal new config", zap.Error(err))
		}

		conf.update()
	})
	viper.WatchConfig()

	return conf, conf.update()
}

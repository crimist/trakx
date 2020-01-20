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

type Config struct {
	Logger *zap.Logger
	Trakx  struct {
		Honey  string `yaml:"honey"`
		Prod   bool   `yaml:"prod"`
		Index  string `yaml:"index"`
		Expvar struct {
			Every int `yaml:"every"`
		} `yaml:"expvar"`
		Pprof struct {
			Port int `yaml:"port"`
		} `yaml:"pprof"`
		Ulimit uint64 `yaml:"ulimit"`
	} `yaml:"trakx"`
	Tracker struct {
		Announce int32 `yaml:"announce"`
		HTTP     struct {
			Enabled      bool `yaml:"enabled"`
			Port         int  `yaml:"port"`
			ReadTimeout  int  `yaml:"readtimeout"`
			WriteTimeout int  `yaml:"writetimeout"`
			Qsize        int  `yaml:"qsize"`
			Accepters    int  `yaml:"accepters"`
			Threads      int  `yaml:"threads"`
		} `yaml:"http"`
		UDP struct {
			Enabled     bool `yaml:"enabled"`
			Port        int  `yaml:"port"`
			CheckConnID bool `yaml:"checkconnid"`
			Threads     int  `yaml:"threads"`
		} `yaml:"udp"`
		Numwant struct {
			Default int32 `yaml:"default"`
			Limit   int32 `yaml:"limit"`
		} `yaml:"numwant"`
	} `yaml:"tracker"`
	Database struct {
		Type   string `yaml:"type"`
		Backup string `yaml:"backup"`
		Peer   struct {
			Filename string `yaml:"filename"`
			Trim     int    `yaml:"trim"`
			Write    int    `yaml:"write"`
			Timeout  int64  `yaml:"timeout"`
		} `yaml:"peer"`
		Conn struct {
			Filename string `yaml:"filename"`
			Trim     int    `yaml:"trim"`
			Timeout  int64  `yaml:"timeout"`
		} `yaml:"conn"`
	} `yaml:"database"`
}

// Loaded checks if the config is loaded or not
func (conf *Config) Loaded() bool {
	// Database.Type is required to run so if it's empty we know that the config isn't loaded
	return conf.Database.Type != ""
}

func (conf *Config) fixFilepaths() error {
	// makes the tild (~) expand out to your home directory

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	conf.Database.Conn.Filename = strings.ReplaceAll(conf.Database.Conn.Filename, "~", home)
	conf.Database.Peer.Filename = strings.ReplaceAll(conf.Database.Peer.Filename, "~", home)
	conf.Trakx.Index = strings.ReplaceAll(conf.Trakx.Index, "~", home)

	return nil
}

// sets the limits loaded in the conf
func (conf *Config) setLimits() error {
	var rLimit syscall.Rlimit

	if conf.Trakx.Ulimit == 0 {
		return nil
	}

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
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
		return err
	}
	return nil
}

func LoadConf(logger *zap.Logger) (*Config, error) {
	conf := new(Config)
	conf.Logger = logger

	// Load from file
	viper.SetConfigType("yaml")
	viper.SetConfigName("trakx")
	viper.AddConfigPath("/usr/local/trakx/")
	viper.AddConfigPath("/app/")
	viper.AddConfigPath(".")
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

		conf.setLimits()
		conf.fixFilepaths()
	})
	viper.WatchConfig()

	conf.setLimits()
	conf.fixFilepaths()

	return conf, nil
}

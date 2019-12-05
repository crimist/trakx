package shared

import (
	"errors"
	"os"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
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
		Announce   int32  `yaml:"announce"`
		StoppedMsg string `yaml:"stoppedmsg"`
		HTTP       struct {
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
			Max     int32 `yaml:"max"`
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

func (conf *Config) Loaded() bool {
	return conf.Database.Type != ""
}

// makes the tild (~) expand out to your home directory
func (conf *Config) fixFilenames() error {
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

func ViperConf(logger *zap.Logger) (*Config, error) {
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
		return nil, errors.New("ReadInConfig failed with error: " + err.Error())
	}

	// Env vars override file
	// Thanks https://github.com/spf13/viper/issues/188#issuecomment-413368673
	viper.AutomaticEnv()
	viper.SetEnvPrefix("trakx")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.Unmarshal(conf); err != nil {
		return nil, errors.New("Unmarshal failed with error: " + err.Error())
	}

	// If $PORT var set override everything for appengines
	viper.BindEnv("app_port", "PORT")
	if appenginePort := viper.GetInt("app_port"); appenginePort != 0 {
		conf.Tracker.HTTP.Port = appenginePort
	}

	// Add watcher
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		conf.Logger.Info("Config changed", zap.String("name", e.Name), zap.Any("op", e.Op))
		if err := viper.Unmarshal(conf); err != nil {
			conf.Logger.Info("New config invalid", zap.Error(err))
		}
		conf.setLimits()
		conf.fixFilenames()
	})

	conf.setLimits()
	conf.fixFilenames()

	return conf, nil
}

package shared

import (
	"os"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Trakx struct {
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
		HTTP struct {
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
		StoppedMsg       string `yaml:"stoppedmsg"`
		AnnounceInterval int32  `yaml:"announce"`
	} `yaml:"tracker"`
	Database struct {
		Peer struct {
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

func ViperConf(logger *zap.Logger) *Config {
	conf := new(Config)

	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	viper.AddConfigPath("$HOME/.trakx/")
	viper.AddConfigPath("/app/")
	viper.AddConfigPath(".")

	viper.BindEnv("app_port", "PORT")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Panic("Failed to read config", zap.Error(err))
	}
	if err := viper.Unmarshal(conf); err != nil {
		logger.Panic("Invalid config", zap.Error(err))
	}

	// If $PORT var set override
	if appenginePort := viper.GetInt("app_port"); appenginePort != 0 {
		conf.Tracker.HTTP.Port = appenginePort
	}

	conf.setLimits()
	conf.fixFilenames()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("Config changed", zap.String("name", e.Name), zap.Any("op", e.Op))
		if err := viper.Unmarshal(conf); err != nil {
			logger.Info("New config invalid", zap.Error(err))
		}
		conf.setLimits()
		conf.fixFilenames()
	})

	return conf
}

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

func (conf *Config) setLimits() error {
	if conf.Trakx.Ulimit == 0 {
		return nil
	}
	var rLimit syscall.Rlimit
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

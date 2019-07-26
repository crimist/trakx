package shared

import (
	"io/ioutil"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	PeerDB PeerDatabase
	Logger *zap.Logger
	Env    Enviroment
	Config struct {
		Trakx struct {
			Prod bool `yaml:"prod"`
		} `yaml:"trakx"`

		Tracker struct {
			HTTP        bool `yaml:"http"`
			Checkconnid bool `yaml:"checkconnid"`
			Ports       struct {
				UDP    int `yaml:"udp"`
				HTTP   int `yaml:"http"`
				Expvar int `yaml:"expvar"`
			} `yaml:"ports"`
			Numwant struct {
				Default int32 `yaml:"default"`
				Max     int32 `yaml:"max"`
			} `yaml:"numwant"`
			StoppedMsg       string `yaml:"stopped"`
			MetricsInterval  int    `yaml:"metrics"`
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
			} `yaml:"conn"`
		} `yaml:"database"`
		TrakxRoot string
	}
)

// LoadConfig loads the yaml config at this projects root
func LoadConfig() {
	data, err := ioutil.ReadFile("trakx.yaml")
	if err != nil {
		Logger.Panic("Failed to load config", zap.Error(err))
	}
	if err = yaml.Unmarshal(data, &Config); err != nil {
		Logger.Panic("Failed to parse config", zap.Error(err))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Panic("os.UserHomeDir() failed", zap.Error(err))
	}

	Config.TrakxRoot = home + "/.trakx"
	Config.Database.Peer.Filename = Config.TrakxRoot + Config.Database.Peer.Filename
	Config.Database.Conn.Filename = Config.TrakxRoot + Config.Database.Conn.Filename
}

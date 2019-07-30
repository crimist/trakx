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

	Config struct {
		Trakx struct {
			Prod   bool   `yaml:"prod"`
			Index  string `yaml:"index"`
			Expvar struct {
				Enabled bool `yaml:"enabled"`
				Every   int  `yaml:"every"`
				Port    int  `yaml:"port"`
			} `yaml:"expvar"`
		} `yaml:"trakx"`
		Tracker struct {
			HTTP struct {
				Enabled bool `yaml:"enabled"`
				Port    int  `yaml:"port"`
			} `yaml:"http"`
			UDP struct {
				Enabled     bool `yaml:"enabled"`
				Port        int  `yaml:"port"`
				CheckConnID bool `yaml:"checkconnid"`
			} `yaml:"udp"`
			Numwant struct {
				Default int32 `yaml:"default"`
				Max     int32 `yaml:"max"`
			} `yaml:"numwant"`
			StoppedMsg       string `yaml:"stoppedmsg"`
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
				Timeout  int64  `yaml:"timeout"`
			} `yaml:"conn"`
		} `yaml:"database"`
	}
)

// LoadConfig loads the yaml config at this projects root
func LoadConfig() {
	root, err := os.UserHomeDir()
	if err != nil {
		panic("os.UserHomeDir() failed: " + err.Error())
	}
	root += "/.trakx/"

	data, err := ioutil.ReadFile(root + "config.yaml")
	if err != nil {
		panic("Failed to read config: " + err.Error())
	}
	if err = yaml.Unmarshal(data, &Config); err != nil {
		panic("Failed to parse config: " + err.Error())
	}

	Config.Trakx.Index = root + Config.Trakx.Index
	Config.Database.Peer.Filename = root + Config.Database.Peer.Filename
	Config.Database.Conn.Filename = root + Config.Database.Conn.Filename
}

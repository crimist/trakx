package shared

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
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

// NewConfig loads "config.yaml" at root
func NewConfig(root string) *Config {
	conf := Config{}
	conf.Load(root)

	return &conf
}

func (conf *Config) Load(root string) {
	if conf == nil {
		panic("conf == nil")
	}

	data, err := ioutil.ReadFile(root + "config.yaml")
	if err != nil {
		panic("Failed to read config: " + err.Error())
	}
	if err = yaml.Unmarshal(data, conf); err != nil {
		panic("Failed to parse config: " + err.Error())
	}

	conf.Trakx.Index = root + conf.Trakx.Index
	conf.Database.Peer.Filename = root + conf.Database.Peer.Filename
	conf.Database.Conn.Filename = root + conf.Database.Conn.Filename
}

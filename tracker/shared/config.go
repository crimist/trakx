package shared

import (
	"io/ioutil"
	"runtime"
	"syscall"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Trakx struct {
		Prod   bool   `yaml:"prod"`
		Index  string `yaml:"index"`
		Expvar struct {
			Every int `yaml:"every"`
			Port  int `yaml:"port"`
		} `yaml:"expvar"`
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

// NewConfig loads "config.yaml" at root
func NewConfig(root string) (*Config, error) {
	conf := Config{}
	if err := conf.Load(root); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (conf *Config) Load(root string) error {
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

	// Set ulimit
	if conf.Trakx.Ulimit == 0 {
		return nil
	}
	var rLimit syscall.Rlimit
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}
	// Bugged rn so OSX can't set above 24000ish
	if runtime.GOOS != "darwin" {
		rLimit.Max = conf.Trakx.Ulimit
		rLimit.Cur = conf.Trakx.Ulimit
	}
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}
	return nil
}

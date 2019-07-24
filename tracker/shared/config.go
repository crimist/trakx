package shared

import (
	"io/ioutil"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var Config struct {
	Production       bool `yaml:"prod"`
	HTTPPort         int    `yaml:"httpport"`
	HTTPTracker             bool    `yaml:"http"`
	UDPPort          int    `yaml:"udp"`
	ExpvarPort       int    `yaml:"expvar"`
	NumwantDefault   int32  `yaml:"numwant"`
	NumwantMax       int32  `yaml:"maxnumwant"`
	StoppedMsg       string `yaml:"bye"`
	UDPTrim          int    `yaml:"udptrim"`
	MetricsInterval  int    `yaml:"metrics"`
	DBWriteInterval  int    `yaml:"write"`
	DBCleanInterval  int    `yaml:"cleaninterval"`
	DBCleanTimeout   int64  `yaml:"cleantimeout"`
	AnnounceInterval int32  `yaml:"announce"`

	Database struct {
		Peer string `yaml:"peer"`
		Conn string `yaml:"conn"`
	} `yaml:"database"`
}

func loadConfig() {
	data, err := ioutil.ReadFile("trakx.yaml")
	if err != nil {
		Logger.Panic("Failed to load config", zap.Error(err))
	}
	if err = yaml.Unmarshal(data, &Config); err != nil {
		Logger.Panic("Failed to parse config", zap.Error(err))
	}
}

var (
	PeerDB         PeerDatabase
	Logger         *zap.Logger
	Env            Enviroment
	UDPCheckConnID bool
)

package config

import (
	"io/ioutil"
	"os"
	"strings"

	_ "github.com/crimist/trakx/embedded/statik"
	"github.com/rakyll/statik/fs"
	"go.uber.org/zap"
)

// generate with `statik -src ./embedded -include "*.html,*.yaml"`

const (
	defaultMessage = "n/a"
)

var (
	IndexData = defaultMessage // http index page string
	DMCAData  = defaultMessage // http dmca page string
)

// loadEmbed loads all the embedded files in the exe and sets up crutial filesystem
func loadEmbed() {
	fileSys, err := fs.New()
	if err != nil {
		Logger.Panic("failed to open statik fs", zap.Error(err))
	}

	// create config if it doesn't exist
	_, err = os.Stat(ConfigDir + "trakx.yaml")
	if os.IsNotExist(err) {
		cfgData, err := fs.ReadFile(fileSys, "/trakx.yaml")
		if err != nil {
			Logger.Error("failed to read embedded config", zap.Error(err))
		}
		err = ioutil.WriteFile(ConfigDir+"trakx.yaml", cfgData, FilePerm)
		if err != nil {
			Logger.Error("failed to write config file", zap.Error(err))
		}
	} else if err != nil {
		Logger.Error("failed to stat config file", zap.Error(err))
	}

	// load HTML
	if indexTmp, err := fs.ReadFile(fileSys, "/index.html"); err != nil {
		Logger.Error("failed to read index file from statik fs", zap.Error(err))
	} else {
		IndexData = string(indexTmp)
	}
	if dmcaTmp, err := fs.ReadFile(fileSys, "/dmca.html"); err != nil {
		Logger.Error("failed to read dmca file from statik fs", zap.Error(err))
	} else {
		DMCAData = string(dmcaTmp)
	}

	// trim whitespace to save bandwidth
	strip := func(data string) string {
		data = strings.ReplaceAll(data, "\t", "")
		data = strings.ReplaceAll(data, "\n", "")
		return data
	}

	IndexData = strip(IndexData)
	DMCAData = strip(DMCAData)
}

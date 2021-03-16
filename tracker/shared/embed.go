package shared

import (
	"io/ioutil"
	"os"
	"strings"

	_ "github.com/crimist/trakx/statik"
	"github.com/crimist/trakx/tracker/paths"
	"github.com/rakyll/statik/fs"
	"go.uber.org/zap"
)

// generate with `statik -src ./install -include "*.html,*.yaml"`

const (
	defaultMessage = "n/a"
)

var (
	IndexData      = defaultMessage
	IndexDataBytes = []byte(defaultMessage)
	DMCAData       = defaultMessage
	DMCADataBytes  = []byte(defaultMessage)
)

// LoadEmbed loads all the embedded files in the exe and sets up crutial filesystem
func LoadEmbed(logger *zap.Logger) {
	fileSys, err := fs.New()
	if err != nil {
		logger.Panic("failed to open statik fs", zap.Error(err))
	}

	// add config if it doesn't exist
	_, err = os.Stat(paths.ConfigDir + "trakx.yaml")
	if os.IsNotExist(err) {
		cfgData, err := fs.ReadFile(fileSys, "/trakx.yaml")
		if err != nil {
			logger.Error("failed to read embedded config", zap.Error(err))
		}
		err = ioutil.WriteFile(paths.ConfigDir+"trakx.yaml", cfgData, paths.FilePerm)
		if err != nil {
			logger.Error("failed to write config file", zap.Error(err))
		}
	} else if err != nil {
		logger.Error("failed to stat config file", zap.Error(err))
	}

	// load HTML
	if IndexDataBytes, err = fs.ReadFile(fileSys, "/index.html"); err != nil {
		logger.Error("failed to read index file from statik fs", zap.Error(err))
	} else {
		IndexData = string(IndexDataBytes)
	}
	if DMCADataBytes, err = fs.ReadFile(fileSys, "/dmca.html"); err != nil {
		logger.Error("failed to read dmca file from statik fs", zap.Error(err))
	} else {
		DMCAData = string(DMCADataBytes)
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

package shared

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	_ "github.com/crimist/trakx/statik"
	"github.com/rakyll/statik/fs"
	"go.uber.org/zap"
)

// generate with `statik -src ./install -include "*.html,*.yaml"`

const (
	// FolderPerm is the default permission mask for folders
	FolderPerm = 0700
	// FilePerm is the default permission mask for files
	FilePerm = 0644

	uninit = "uninitialized, you should never see this"
)

var (
	IndexData      = uninit
	IndexDataBytes = []byte(uninit)
	DMCAData       = uninit
	DMCADataBytes  = []byte(uninit)

	ConfigDir string
	CacheDir  string
)

// LoadEmbed loads all the embedded files in the exe and sets up crutial filesystem
func LoadEmbed(logger *zap.Logger) {
	fileSys, err := fs.New()
	if err != nil {
		logger.Panic("failed to open statik fs", zap.Error(err))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("failed to get user home dir", zap.Error(err))
	}

	ConfigDir = home + "/.config/trakx/"
	CacheDir = home + "/.cache/trakx/"

	oldmask := syscall.Umask(0)
	err = os.MkdirAll(CacheDir, FolderPerm)
	if err != nil {
		logger.Error("failed to create cache dir", zap.Error(err))
	}
	err = os.MkdirAll(ConfigDir, FolderPerm)
	if err != nil {
		logger.Error("failed to create config dir", zap.Error(err))
	}
	syscall.Umask(oldmask)

	// add config if it doesn't exist
	_, err = os.Stat(ConfigDir + "trakx.yaml")
	if os.IsNotExist(err) {
		cfgData, err := fs.ReadFile(fileSys, "/trakx.yaml")
		if err != nil {
			logger.Error("failed to read embedded config", zap.Error(err))
		}
		err = ioutil.WriteFile(ConfigDir+"trakx.yaml", cfgData, FilePerm)
		if err != nil {
			logger.Error("failed to write config file", zap.Error(err))
		}
	} else if err != nil {
		logger.Error("failed to stat config file", zap.Error(err))
	}

	if IndexDataBytes, err = fs.ReadFile(fileSys, "/index.html"); err != nil {
		logger.Error("failed to read index file from statik fs", zap.Error(err))
	}
	if DMCADataBytes, err = fs.ReadFile(fileSys, "/dmca.html"); err != nil {
		logger.Error("failed to read dmca file from statik fs", zap.Error(err))
	}

	IndexData = string(IndexDataBytes)
	DMCAData = string(DMCADataBytes)

	// trim whitespace to save bytes
	strip := func(data string) string {
		data = strings.ReplaceAll(data, "\t", "")
		data = strings.ReplaceAll(data, "\n", "")
		return data
	}

	IndexData = strip(IndexData)
	DMCAData = strip(DMCAData)
}

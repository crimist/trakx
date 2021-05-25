package config

import (
	"os"
	"syscall"

	"go.uber.org/zap"
)

// automatically sets up the filesys when imported

const (
	// FolderPerm is the default permission mask for folders
	FolderPerm = 0700
	// FilePerm is the default permission mask for files
	FilePerm = 0644
)

var (
	// ConfigDir stores the absolute path of the config directory
	ConfigDir string
	// CacheDir stores the absolute path of the cache directory
	CacheDir string
)

func initPaths() {
	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Panic("failed to get user home dir", zap.Error(err))
	}

	ConfigDir = home + "/.config/trakx/"
	CacheDir = home + "/.cache/trakx/"

	syscall.Umask(0)

	err = os.MkdirAll(CacheDir, FolderPerm)
	if err != nil {
		Logger.Panic("failed to create cache dir", zap.Error(err))
	}
	err = os.MkdirAll(ConfigDir, FolderPerm)
	if err != nil {
		Logger.Panic("failed to create config dir", zap.Error(err))
	}
}

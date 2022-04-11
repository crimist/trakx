package config

import (
	"os"
	"syscall"

	"go.uber.org/zap"
)

// automatically sets up the filesys when imported

const (
	// FolderPerm holds the default permission mask for folders
	FolderPerm = 0700
	// FilePerm holds the default permission mask for files
	FilePerm = 0644
)

var (
	// configPath stores the absolute path of the config directory
	configPath string
	// CachePath stores the absolute path of the cache directory
	CachePath string
)

func initPaths() {
	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Panic("failed to get user home dir", zap.Error(err))
	}

	configPath = home + "/.config/trakx/"
	CachePath = home + "/.cache/trakx/"

	syscall.Umask(0)

	err = os.MkdirAll(CachePath, FolderPerm)
	if err != nil {
		Logger.Panic("failed to create cache dir", zap.Error(err))
	}
	err = os.MkdirAll(configPath, FolderPerm)
	if err != nil {
		Logger.Panic("failed to create config dir", zap.Error(err))
	}
}

package config

import (
	"os"
	"syscall"

	"go.uber.org/zap"
)

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

// initDirectories attempts to create cache and config directories
func initDirectories() {
	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Warn("failed to get users home directory", zap.Error(err))
		return
	}

	configPath = home + "/.config/trakx/"
	CachePath = home + "/.cache/trakx/"

	syscall.Umask(0)

	err = os.MkdirAll(CachePath, FolderPerm)
	if err != nil {
		Logger.Warn("failed to create cache directory", zap.Error(err))
	}
	err = os.MkdirAll(configPath, FolderPerm)
	if err != nil {
		Logger.Warn("failed to create config directory", zap.Error(err))
	}
}

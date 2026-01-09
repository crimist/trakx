package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//go:embed embedded/*
var embeddedFS embed.FS

func installDefaultConfig(path string) error {
	syscall.Umask(0)

	_, err := os.Stat(path)

	if err == nil {
		zap.L().Debug("configuration file already exists, skipping installation", zap.String("path", path))
		return nil
	} else if !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to stat config file "+path)
	}

	configurationContents, err := embeddedFS.ReadFile("embedded/trakx.yaml")
	if err != nil {
		return errors.Wrap(err, "failed to read config from embedded FS")
	}

	if err = os.MkdirAll(filepath.Dir(path), defaultFolderPermission); err != nil {
		zap.L().Warn("failed to create configuration directory", zap.Error(err), zap.String("directory", filepath.Dir(path)))
	}

	if err = os.WriteFile(path, configurationContents, defaultFilePermission); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to write configuration file '%s'", path))
	}

	zap.L().Debug("installed default configuration", zap.String("path", path))

	return nil
}

func DumpDefaultConfig() error {
	configurationContents, err := embeddedFS.ReadFile("embedded/trakx.yaml")
	if err != nil {
		return errors.Wrap(err, "failed to read config from embedded FS")
	}

	fmt.Print(string(configurationContents))

	return nil
}

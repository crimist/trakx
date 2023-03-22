package config

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//go:embed embedded/*
var embeddedFS embed.FS

func writeEmbeddedConfig(path string) error {
	syscall.Umask(0)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		configData, err := embeddedFS.ReadFile("embedded/trakx.yaml")
		if err != nil {
			return errors.Wrap(err, "failed to read embedded FS")
		}

		if err = os.MkdirAll(filepath.Dir(path), folderPerm); err != nil {
			zap.L().Warn("failed to create config directory", zap.Error(err))
		}

		if err = os.WriteFile(path, configData, filePerm); err != nil {
			return errors.Wrap(err, "failed to write configuration file to "+path)
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to stat config file "+path)
	}

	return nil
}

type EmbeddedCache map[string]string

func stripNewlineTabs(data string) string {
	data = strings.ReplaceAll(data, "\t", "")
	data = strings.ReplaceAll(data, "\n", "")
	return data
}

func GenerateEmbeddedCache() (EmbeddedCache, error) {
	dir, err := embeddedFS.ReadDir("embedded")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read embed directory to populate cache")
	}

	cache := make(EmbeddedCache, len(dir))

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// don't expose configuration
		if filename == "trakx.yaml" {
			continue
		}

		data, err := embeddedFS.ReadFile("embedded/" + filename)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %v from embedded", "embedded/"+filename)
		}

		// trim data from html files to save bandwidth
		dataStr := string(data)
		if strings.HasSuffix(filename, ".html") {
			dataStr = stripNewlineTabs(dataStr)
		}

		zap.L().Debug("adding file to embedded cache", zap.String("filename", filename))
		cache["/"+filename] = dataStr
	}

	// if index exists copy to /
	if indexData, ok := cache["/index.html"]; ok {
		cache["/"] = indexData
	}

	return cache, nil
}

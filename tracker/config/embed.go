package config

import (
	"embed"
	"os"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//go:embed embedded/*
var embeddedFileSystem embed.FS

func generateConfig() {
	if _, err := os.Stat(configPath + "trakx.yaml"); os.IsNotExist(err) {
		configData, err := embeddedFileSystem.ReadFile("embedded/trakx.yaml")
		if err != nil {
			Logger.Error("failed to read embedded config", zap.Error(err))
			return
		}
		err = os.WriteFile(configPath+"trakx.yaml", configData, FilePerm)
		if err != nil {
			Logger.Error("failed to write config file", zap.Error(err))
		}
	} else if err != nil {
		Logger.Error("failed to stat config file", zap.Error(err))
	}
}

type EmbeddedCache map[string]string

func stripNewlineTabs(data string) string {
	data = strings.ReplaceAll(data, "\t", "")
	data = strings.ReplaceAll(data, "\n", "")
	return data
}

func GenerateEmbeddedCache() (EmbeddedCache, error) {
	dir, err := embeddedFileSystem.ReadDir("embedded")
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

		data, err := embeddedFileSystem.ReadFile("embedded/" + filename)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %v from embedded", "embedded/"+filename)
		}

		// trim data from html files to save bandwidth
		dataStr := string(data)
		if strings.HasSuffix(filename, ".html") {
			dataStr = stripNewlineTabs(dataStr)
		}

		Logger.Debug("adding file to embedded cache", zap.String("filename", filename))
		cache["/"+filename] = dataStr
	}

	// if index exists copy to /
	if indexData, ok := cache["/index.html"]; ok {
		cache["/"] = indexData
	}

	return cache, nil
}

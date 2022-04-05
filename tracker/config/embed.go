package config

import (
	"embed"
	_ "embed"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//go:embed embedded/*
var Embedded embed.FS

type EmbeddedCache map[string]string

func initEmbedded() {
	// create config if it doesn't exist
	_, err := os.Stat(ConfigDir + "trakx.yaml")
	if os.IsNotExist(err) {
		cfgData, err := Embedded.ReadFile("embedded/trakx.yaml")
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
}

func GenerateEmbeddedCache() (EmbeddedCache, error) {
	strip := func(data string) string {
		data = strings.ReplaceAll(data, "\t", "")
		data = strings.ReplaceAll(data, "\n", "")
		return data
	}

	dir, err := Embedded.ReadDir("embedded")
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

		data, err := Embedded.ReadFile("embedded/" + filename)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %v from embedded", "embedded/"+filename)
		}

		// trim data from html files to save bandwidth
		dataStr := string(data)
		if strings.HasSuffix(filename, ".html") {
			dataStr = strip(dataStr)
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

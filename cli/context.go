package main

import (
	"io"

	"github.com/crimist/trakx/config"
)

type GlobalOptions struct {
	ConfigPath string
	Format     string
	ShowHelp   bool
}

type Context struct {
	Global GlobalOptions
	Stdout io.Writer
	Stderr io.Writer
	config *config.Configuration
}

func (c *Context) LoadConfig() (*config.Configuration, error) {
	if c.config != nil {
		return c.config, nil
	}

	conf, err := config.Load(config.LoadOptions{
		Path: c.Global.ConfigPath,
	})
	if err != nil {
		return nil, err
	}
	c.config = conf
	return conf, nil
}

func (c *Context) ResolvedConfigPath() (string, error) {
	if c.Global.ConfigPath != "" {
		return c.Global.ConfigPath, nil
	}
	return config.DefaultPath()
}

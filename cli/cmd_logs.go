package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	logPollInterval = 500 * time.Millisecond
)

func newLogsCommand() *Command {
	flags, help := newFlagSet("logs")
	follow := flags.Bool("follow", false, "Follow log output")
	flags.BoolVar(follow, "f", false, "Follow log output")

	cmd := &Command{
		Name:     "logs",
		Short:    "Print most recent log file",
		Usage:    "trakx logs [--follow]",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}

			path, err := findLatestLogFile(conf.Cache)
			if err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(ctx.Stdout, file); err != nil {
				return err
			}

			if *follow {
				return followFile(file, ctx.Stdout)
			}

			return nil
		},
	}
	return cmd
}

func findLatestLogFile(cacheDir string) (string, error) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", err
	}

	var latestPath string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isLogFile(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return "", err
		}

		modTime := info.ModTime()
		if latestPath == "" || modTime.After(latestTime) {
			latestTime = modTime
			latestPath = filepath.Join(cacheDir, name)
		}
	}

	if latestPath == "" {
		return "", errors.New("no log files found")
	}

	return latestPath, nil
}

func isLogFile(name string) bool {
	if len(name) < len("trakx_.log") {
		return false
	}
	if name[:6] != "trakx_" {
		return false
	}
	return filepath.Ext(name) == ".log"
}

func followFile(file *os.File, out io.Writer) error {
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(logPollInterval)
				continue
			}
			return err
		}
	}
}

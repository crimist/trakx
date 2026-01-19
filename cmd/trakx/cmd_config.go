package main

import (
	"fmt"

	"github.com/crimist/trakx/internal/config"
)

func newConfigCommand() *Command {
	cmd := &Command{
		Name:  "config",
		Short: "Manage configuration",
		Usage: "trakx config <subcommand>",
	}
	cmd.Add(newConfigDumpCommand())
	cmd.Add(newConfigPathCommand())
	cmd.Add(newConfigValidateCommand())
	return cmd
}

func newConfigDumpCommand() *Command {
	flags, help := newFlagSet("config dump")
	cmd := &Command{
		Name:     "dump",
		Short:    "Print default config",
		Usage:    "trakx config dump",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			return config.DumpDefaultConfig()
		},
	}
	return cmd
}

func newConfigPathCommand() *Command {
	flags, help := newFlagSet("config path")
	cmd := &Command{
		Name:     "path",
		Short:    "Print resolved config path",
		Usage:    "trakx config path",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			path, err := ctx.ResolvedConfigPath()
			if err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, path)
			return nil
		},
	}
	return cmd
}

func newConfigValidateCommand() *Command {
	flags, help := newFlagSet("config validate")
	cmd := &Command{
		Name:     "validate",
		Short:    "Validate config",
		Usage:    "trakx config validate",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			if _, err := ctx.LoadConfig(); err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, "ok")
			return nil
		},
	}
	return cmd
}

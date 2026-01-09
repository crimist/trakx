package main

import (
	"fmt"
	"runtime/debug"
)

func newVersionCommand() *Command {
	flags, help := newFlagSet("version")
	cmd := &Command{
		Name:     "version",
		Short:    "Print build info",
		Usage:    "trakx version",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			version := "dev"
			if info, ok := debug.ReadBuildInfo(); ok {
				if info.Main.Version != "" && info.Main.Version != "(devel)" {
					version = info.Main.Version
				}
			}
			fmt.Fprintln(ctx.Stdout, version)
			return nil
		},
	}
	return cmd
}

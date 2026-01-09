package main

import "fmt"

func newStartCommand() *Command {
	flags, help := newFlagSet("start")
	importPath := flags.String("import", "", "Import snapshot from file (start only, no stdin)")
	cmd := &Command{
		Name:     "start",
		Short:    "Start tracker in background",
		Usage:    "trakx start [--import <file>]",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			controller := newDaemonController(conf)
			if *importPath == "-" {
				return fmt.Errorf("start does not support stdin import; use a file path")
			}
			if err := controller.Start(ctx.Global, *importPath); err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, "started")
			return nil
		},
	}
	return cmd
}

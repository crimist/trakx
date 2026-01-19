package main

import "fmt"

func newStopCommand() *Command {
	flags, help := newFlagSet("stop")
	cmd := &Command{
		Name:     "stop",
		Short:    "Stop background process",
		Usage:    "trakx stop",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			controller := newDaemonController(conf)
			fmt.Fprintln(ctx.Stdout, "stopping...")
			if err := controller.Stop(ctx.Stdout); err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, "stopped")
			return nil
		},
	}
	return cmd
}

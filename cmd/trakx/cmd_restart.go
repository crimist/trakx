package main

import "fmt"

func newRestartCommand() *Command {
	flags, help := newFlagSet("restart")
	cmd := &Command{
		Name:     "restart",
		Aliases:  []string{"reboot"},
		Short:    "Stop then start",
		Usage:    "trakx restart",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			controller := newDaemonController(conf)
			fmt.Fprintln(ctx.Stdout, "restarting...")
			if err := controller.Stop(ctx.Stdout); err != nil {
				return err
			}
			if err := controller.Start(ctx.Global, ""); err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, "restarted")
			return nil
		},
	}
	return cmd
}

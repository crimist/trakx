package main

import (
	"fmt"
	"os"
	"syscall"
)

func newPidCommand() *Command {
	cmd := &Command{
		Name:  "pid",
		Short: "Inspect and manage the pid file",
		Usage: "trakx pid <subcommand>",
	}
	cmd.Add(newPidShowCommand())
	cmd.Add(newPidClearCommand())
	cmd.Add(newPidAliveCommand())
	cmd.Add(newPidPathCommand())
	return cmd
}

func newPidShowCommand() *Command {
	flags, help := newFlagSet("pid show")
	cmd := &Command{
		Name:     "show",
		Short:    "Print pid from pid file",
		Usage:    "trakx pid show",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			pidFile := newProcessIDFile(conf.PIDPath())
			pid, err := pidFile.Read()
			if err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, pid)
			return nil
		},
	}
	return cmd
}

func newPidClearCommand() *Command {
	flags, help := newFlagSet("pid clear")
	cmd := &Command{
		Name:     "clear",
		Short:    "Clear the pid file",
		Usage:    "trakx pid clear",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			pidFile := newProcessIDFile(conf.PIDPath())
			if err := pidFile.Clear(); err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, "pid file cleared")
			return nil
		},
	}
	return cmd
}

func newPidAliveCommand() *Command {
	flags, help := newFlagSet("pid alive")
	cmd := &Command{
		Name:     "alive",
		Short:    "Check if pid is alive",
		Usage:    "trakx pid alive",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			pidFile := newProcessIDFile(conf.PIDPath())
			pid, err := pidFile.Read()
			if err != nil {
				return err
			}
			if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {
				fmt.Fprintln(ctx.Stdout, "alive")
			} else if os.IsNotExist(err) {
				fmt.Fprintln(ctx.Stdout, "dead")
			} else if err.Error() == "no such process" {
				fmt.Fprintln(ctx.Stdout, "dead")
			} else {
				fmt.Fprintln(ctx.Stdout, "dead")
			}
			return nil
		},
	}
	return cmd
}

func newPidPathCommand() *Command {
	flags, help := newFlagSet("pid path")
	cmd := &Command{
		Name:     "path",
		Short:    "Print pid file path",
		Usage:    "trakx pid path",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, conf.PIDPath())
			return nil
		},
	}
	return cmd
}

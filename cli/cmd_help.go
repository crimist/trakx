package main

import "fmt"

func newHelpCommand(root *Command) *Command {
	flags, help := newFlagSet("help")
	cmd := &Command{
		Name:     "help",
		Short:    "Help about any command",
		Usage:    "trakx help [command]",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			if len(args) == 0 {
				root.PrintHelp(ctx)
				return nil
			}

			if target, ok := root.FindPath(args); ok {
				target.PrintHelp(ctx)
				return nil
			}

			return fmt.Errorf("unknown command: %s", args[0])
		},
	}
	return cmd
}

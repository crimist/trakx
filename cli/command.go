package main

import (
	"flag"
)

type Command struct {
	Name        string
	Aliases     []string
	Short       string
	Long        string
	Usage       string
	Flags       *flag.FlagSet
	HelpFlag    *bool
	Subcommands []*Command
	Run         func(*Context, []string) error
	Hidden      bool
}

func (c *Command) Add(cmd *Command) {
	c.Subcommands = append(c.Subcommands, cmd)
}

func (c *Command) Find(name string) *Command {
	for _, cmd := range c.Subcommands {
		if cmd.Name == name {
			return cmd
		}
		for _, alias := range cmd.Aliases {
			if alias == name {
				return cmd
			}
		}
	}
	return nil
}

func (c *Command) FindPath(path []string) (*Command, bool) {
	current := c
	for _, part := range path {
		next := current.Find(part)
		if next == nil {
			return nil, false
		}
		current = next
	}
	return current, true
}

func (c *Command) Execute(ctx *Context, args []string) error {
	if c.Flags != nil {
		c.Flags.SetOutput(ctx.Stderr)
		c.Flags.Usage = func() {
			c.PrintHelp(ctx)
		}
		if err := c.Flags.Parse(args); err != nil {
			return err
		}
		if c.HelpFlag != nil && *c.HelpFlag {
			c.PrintHelp(ctx)
			return nil
		}
		args = c.Flags.Args()
	}

	if len(args) > 0 {
		if sub := c.Find(args[0]); sub != nil {
			return sub.Execute(ctx, args[1:])
		}
	}

	if c.Run == nil {
		c.PrintHelp(ctx)
		return nil
	}

	return c.Run(ctx, args)
}

func newFlagSet(name string) (*flag.FlagSet, *bool) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	help := fs.Bool("help", false, "Show help")
	fs.BoolVar(help, "h", false, "Show help")
	return fs, help
}

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/crimist/trakx/internal/config"
)

func (c *Command) PrintHelp(ctx *Context) {
	if c.Name == "trakx" {
		printRootHelp(ctx, c)
		return
	}
	printCommandHelp(ctx, c)
}

func printRootHelp(ctx *Context, root *Command) {
	fmt.Fprintln(ctx.Stderr, "Trakx is a BitTorrent tracker.")
	fmt.Fprintln(ctx.Stderr, "")
	fmt.Fprintln(ctx.Stderr, "Usage:")
	fmt.Fprintln(ctx.Stderr, "  trakx [global options] <command> [command options]")
	fmt.Fprintln(ctx.Stderr, "")
	fmt.Fprintln(ctx.Stderr, "Commands:")
	printCommandList(ctx, root.Subcommands)
	fmt.Fprintln(ctx.Stderr, "")
	fmt.Fprintln(ctx.Stderr, "Global options:")
	printGlobalOptions(ctx)
	fmt.Fprintln(ctx.Stderr, "")
	fmt.Fprintln(ctx.Stderr, "Config values can also be overridden with TRAKX_* environment variables.")
	fmt.Fprintln(ctx.Stderr, "Example: TRAKX_LOGLEVEL=debug trakx run")
	fmt.Fprintln(ctx.Stderr, "")
	fmt.Fprintln(ctx.Stderr, "Run 'trakx help <command>' for more information about a command.")
}

func printCommandHelp(ctx *Context, cmd *Command) {
	fmt.Fprintf(ctx.Stderr, "Usage:\n  %s\n\n", cmd.Usage)
	if cmd.Long != "" {
		fmt.Fprintln(ctx.Stderr, cmd.Long)
		fmt.Fprintln(ctx.Stderr, "")
	} else if cmd.Short != "" {
		fmt.Fprintln(ctx.Stderr, cmd.Short)
		fmt.Fprintln(ctx.Stderr, "")
	}

	if len(cmd.Subcommands) > 0 {
		fmt.Fprintln(ctx.Stderr, "Subcommands:")
		printCommandList(ctx, cmd.Subcommands)
		fmt.Fprintln(ctx.Stderr, "")
	}

	if cmd.Flags != nil {
		fmt.Fprintln(ctx.Stderr, "Options:")
		cmd.Flags.SetOutput(ctx.Stderr)
		cmd.Flags.PrintDefaults()
		fmt.Fprintln(ctx.Stderr, "")
	}
}

func printCommandList(ctx *Context, cmds []*Command) {
	var visible []*Command
	for _, cmd := range cmds {
		if !cmd.Hidden {
			visible = append(visible, cmd)
		}
	}

	sort.Slice(visible, func(i, j int) bool {
		return visible[i].Name < visible[j].Name
	})

	max := 0
	for _, cmd := range visible {
		if len(cmd.Name) > max {
			max = len(cmd.Name)
		}
	}

	for _, cmd := range visible {
		padding := strings.Repeat(" ", max-len(cmd.Name)+2)
		fmt.Fprintf(ctx.Stderr, "  %s%s%s\n", cmd.Name, padding, cmd.Short)
	}
}

func printGlobalOptions(ctx *Context) {
	defaultPath := "<auto>"
	if path, err := config.DefaultPath(); err == nil {
		defaultPath = path
	}

	fmt.Fprintf(ctx.Stderr, "  --config <path>      Path to config file (default: %s)\n", defaultPath)
	fmt.Fprintln(ctx.Stderr, "  --format <fmt>       Output format for status/output commands: text|json")
	fmt.Fprintln(ctx.Stderr, "  -h, --help           Show help")
}

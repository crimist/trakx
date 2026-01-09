package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	ctx := &Context{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	root := NewRootCommand()

	opts, rest, err := parseGlobal(os.Args[1:])
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err)
		fmt.Fprintln(ctx.Stderr, "Run 'trakx help' for usage.")
		os.Exit(2)
	}
	ctx.Global = opts

	if ctx.Global.ShowHelp {
		if len(rest) == 0 {
			root.PrintHelp(ctx)
			return
		}
		if target, ok := root.FindPath(rest); ok {
			target.PrintHelp(ctx)
			return
		}
		fmt.Fprintf(ctx.Stderr, "unknown command: %s\n\n", rest[0])
		root.PrintHelp(ctx)
		os.Exit(2)
	}
	if len(rest) == 0 {
		root.PrintHelp(ctx)
		return
	}

	if rest[0] == "help" {
		if err := root.Execute(ctx, rest); err != nil {
			fmt.Fprintln(ctx.Stderr, err)
			os.Exit(2)
		}
		return
	}

	cmd := root.Find(rest[0])
	if cmd == nil {
		fmt.Fprintf(ctx.Stderr, "unknown command: %s\n\n", rest[0])
		root.PrintHelp(ctx)
		os.Exit(2)
	}

	if err := cmd.Execute(ctx, rest[1:]); err != nil {
		fmt.Fprintln(ctx.Stderr, err)
		os.Exit(1)
	}
}

func parseGlobal(args []string) (GlobalOptions, []string, error) {
	opts := GlobalOptions{Format: "text"}
	rest := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			opts.ShowHelp = true
		case strings.HasPrefix(arg, "--config="):
			opts.ConfigPath = strings.TrimPrefix(arg, "--config=")
		case arg == "--config":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("missing value for --config")
			}
			opts.ConfigPath = args[i+1]
			i++
		case strings.HasPrefix(arg, "--format="):
			opts.Format = strings.TrimPrefix(arg, "--format=")
		case arg == "--format":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("missing value for --format")
			}
			opts.Format = args[i+1]
			i++
		default:
			rest = append(rest, arg)
		}
	}

	return opts, rest, nil
}

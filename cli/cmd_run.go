package main

import (
	"io"
	"os"

	"github.com/crimist/trakx/cmd"
)

func newRunCommand() *Command {
	flags, help := newFlagSet("run")
	importPath := flags.String("import", "", "Import snapshot from file (use '-' for stdin)")
	cmd := &Command{
		Name:     "run",
		Aliases:  []string{"execute"},
		Short:    "Run tracker in foreground (blocks)",
		Usage:    "trakx run [--import <file>|-]",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}

			var importReader io.Reader
			if *importPath != "" {
				if *importPath == "-" {
					importReader = os.Stdin
				} else {
					file, err := os.Open(*importPath)
					if err != nil {
						return err
					}
					importReader = file
				}
			}

			cmd.RunWithOptions(conf, cmd.RunOptions{
				ImportReader: importReader,
			})
			return nil
		},
	}
	return cmd
}

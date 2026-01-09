package main

import (
	"fmt"
	"io"
	"os"

	"github.com/crimist/trakx/cmd"
)

func newBackupRootCommand() *Command {
	cmd := &Command{
		Name:  "backup",
		Short: "Manage DB backups",
		Usage: "trakx backup <subcommand>",
	}
	cmd.Add(exportSubcommand())
	cmd.Add(importSubcommand())
	return cmd
}

func exportSubcommand() *Command {
	flags, help := newFlagSet("backup export")
	outPath := flags.String("out", "", "Write backup to file (default: stdout)")

	cmd := &Command{
		Name:     "export",
		Short:    "Stream DB backup",
		Usage:    "trakx backup export [--out <file>]",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}

			var out io.Writer = ctx.Stdout
			var file *os.File
			if *outPath != "" {
				file, err = os.Create(*outPath)
				if err != nil {
					return err
				}
				defer file.Close()
				out = file
			}

			if err := cmd.ExportBackup(conf, out); err != nil {
				return err
			}
			if *outPath != "" {
				fmt.Fprintln(ctx.Stdout, "backup exported")
			}
			return nil
		},
	}
	return cmd
}

func importSubcommand() *Command {
	flags, help := newFlagSet("backup import")
	inPath := flags.String("in", "", "Read backup from file (default: stdin)")

	cmd := &Command{
		Name:     "import",
		Short:    "Load DB backup",
		Usage:    "trakx backup import [--in <file>]",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}

			var in io.Reader = os.Stdin
			var file *os.File
			if *inPath != "" {
				file, err = os.Open(*inPath)
				if err != nil {
					return err
				}
				defer file.Close()
				in = file
			}

			if err := cmd.ImportBackup(conf, in); err != nil {
				return err
			}
			if *inPath != "" {
				fmt.Fprintln(ctx.Stdout, "backup imported")
			}
			return nil
		},
	}
	return cmd
}

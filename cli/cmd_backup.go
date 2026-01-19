package main

import (
	"fmt"
	"os"

	"github.com/crimist/trakx/backup"
	"github.com/crimist/trakx/internal/pidfile"
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

			mgr := backup.NewManager(backup.Config{
				BackupFilePath: conf.DB.Backup.Path,
				PIDFile:        pidfile.New(conf.PIDPath()),
				CacheDir:       conf.Cache,
			})

			var dest backup.Destination
			if *outPath != "" {
				dest = backup.NewFileDestination(*outPath)
			} else {
				dest = backup.NewStreamDestination(ctx.Stdout, "stdout")
			}

			if err := mgr.Export(dest); err != nil {
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

			mgr := backup.NewManager(backup.Config{
				BackupFilePath: conf.DB.Backup.Path,
				PIDFile:        pidfile.New(conf.PIDPath()),
				CacheDir:       conf.Cache,
			})

			var source backup.Source
			if *inPath != "" {
				source = backup.NewFileSource(*inPath)
			} else {
				source = backup.NewStreamSource(os.Stdin, "stdin")
			}

			if err := mgr.Import(source); err != nil {
				return err
			}

			fmt.Fprintln(ctx.Stdout, "backup imported")
			return nil
		},
	}
	return cmd
}

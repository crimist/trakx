package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type statusReport struct {
	PIDFileExists bool `json:"pid_file_exists"`
	ProcessAlive  bool `json:"process_alive"`
	HeartbeatOK   bool `json:"heartbeat_ok"`
}

func newStatusCommand() *Command {
	flags, help := newFlagSet("status")
	cmd := &Command{
		Name:     "status",
		Short:    "Show daemon status",
		Usage:    "trakx status",
		Flags:    flags,
		HelpFlag: help,
		Run: func(ctx *Context, args []string) error {
			conf, err := ctx.LoadConfig()
			if err != nil {
				return err
			}
			controller := newDaemonController(conf)
			pidFileExists, processAlive, heartbeat := controller.Status()

			report := statusReport{
				PIDFileExists: pidFileExists,
				ProcessAlive:  processAlive,
				HeartbeatOK:   heartbeat,
			}

			switch ctx.Global.Format {
			case "text":
				printStatusText(ctx, report)
				return nil
			case "json":
				enc := json.NewEncoder(ctx.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			default:
				return errors.New("invalid format: use text or json")
			}
		},
	}
	return cmd
}

func printStatusText(ctx *Context, report statusReport) {
	printStatusLine(ctx, "pid file", report.PIDFileExists)
	printStatusLine(ctx, "process", report.ProcessAlive)
	printStatusLine(ctx, "heartbeat", report.HeartbeatOK)
}

func printStatusLine(ctx *Context, label string, ok bool) {
	status := "fail"
	if ok {
		status = "ok"
	}
	fmt.Fprintf(ctx.Stdout, "%s: %s\n", label, status)
}

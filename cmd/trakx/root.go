package main

func NewRootCommand() *Command {
	root := &Command{
		Name:  "trakx",
		Short: "BitTorrent tracker",
		Usage: "trakx [global options] <command> [command options]",
	}

	root.Add(newRunCommand())
	root.Add(newStartCommand())
	root.Add(newStopCommand())
	root.Add(newRestartCommand())
	root.Add(newStatusCommand())
	root.Add(newLogsCommand())
	root.Add(newPidCommand())
	root.Add(newConfigCommand())
	root.Add(newBackupRootCommand())
	root.Add(newVersionCommand())
	root.Add(newHelpCommand(root))

	return root
}

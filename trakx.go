package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/Syc0x00/Trakx/tracker"
)

var (
	trakxPidFile = "/var/run/trakx.pid"
	trakxLogFile = "/var/log/trakx.log"
)

const (
	trakxPidPerms = 644
	trakxLogPerms = 644
)

func readPid() int {
	data, err := ioutil.ReadFile(trakxPidFile)
	if os.IsNotExist(err) || string(data) == "" {
		return -1
	} else if err != nil {
		panic(err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		panic(err)
	}
	return pid
}

func writePid(pid int) {
	data := []byte(fmt.Sprintf("%d", pid))
	if err := ioutil.WriteFile(trakxPidFile, data, trakxPidPerms); err != nil {
		panic(err)
	}
}

func clearPid() {
	pidFile, err := os.OpenFile(trakxPidFile, os.O_CREATE|os.O_RDWR, trakxPidPerms)
	if err != nil {
		panic(err)
	}
	pidFile.Truncate(0)
	pidFile.Seek(0, 0)
}

func getProcess() *os.Process {
	pid := readPid()
	if pid == -1 {
		fmt.Fprint(os.Stderr, "Trakx isn't running\n")
		return nil
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		panic(err)
	}
	return process
}

func printHelp(arg string) {
	if arg != "" {
		fmt.Fprintf(os.Stderr, "Invalid argument: \"%s\"\n\n", arg)
	}
	help := "Trakx commands:\n"
	help += fmt.Sprintf("  %-12s Runs Trakx (doesn't return)\n", "run")
	help += fmt.Sprintf("  %-12s Starts Trakx as a service\n", "start")
	help += fmt.Sprintf("  %-12s Stops Trakx service\n", "stop")
	help += fmt.Sprintf("  %-12s Restarts Trakx service\n", "restart")
	help += fmt.Sprintf("  %-12s Reloads the Trakx config\n", "reload")
	help += "Usage:\n"
	help += fmt.Sprintf("  %s <command>\n", os.Args[0])
	help += "Example:\n"
	help += fmt.Sprintf("  %s run\n", os.Args[0])

	fmt.Print(help)
}

func start() {
	logFile, err := os.OpenFile(trakxLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, trakxLogPerms)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	cmd := exec.Command(os.Args[0], "run")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		panic(err)
	}
}

func stop() {
	process := getProcess()
	if process == nil {
		os.Exit(-1)
	}
	process.Signal(tracker.SigStop)
	process.Release()

	// process.Wait() fails on most systems since it's not a child
	// As such we just have to keep checking if the process died

	time.Sleep(1 * time.Second)
	var err error
	pid := readPid()
	for ; err == nil; err = syscall.Kill(pid, syscall.Signal(0)) {
		fmt.Println("Waiting for death....")
		time.Sleep(50 * time.Millisecond)
	}
	if err.Error() != "no such process" {
		panic(err)
	}

	clearPid()
}

func main() {
	if len(os.Args) < 2 {
		printHelp("")
		return
	}

	switch os.Args[1] {
	case "run":
		if readPid() != -1 {
			fmt.Fprint(os.Stderr, "Trakx is already running\n")
			os.Exit(-1)
		}
		writePid(os.Getpid())
		tracker.Run()
	case "start":
		if readPid() != -1 {
			fmt.Fprint(os.Stderr, "Trakx is already running\n")
			os.Exit(-1)
		}
		fmt.Println("starting...")
		start()
		fmt.Println("started!")
	case "stop":
		fmt.Println("stopping...")
		stop()
		fmt.Println("stopped!")
	case "restart", "reboot":
		fmt.Println("rebooting...")
		stop()
		start()
		fmt.Println("rebooted!")
	case "reload":
		fmt.Println("reloading...")
		process := getProcess()
		if process == nil {
			os.Exit(-1)
		}
		process.Signal(tracker.SigReload)
		process.Release()
		fmt.Println("reloaded")
	default:
		printHelp(os.Args[1])
	}
}

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

const (
	trakxPerms = 0740
)

type Files struct {
	root string
	pid  string
	log  string
}

var files Files

func (f *Files) init() {
	oldMask := syscall.Umask(0)

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	f.root = home + "/.trakx"
	f.pid = f.root + "/trakx.pid"
	f.log = f.root + "/trakx.log"

	if err := os.MkdirAll(f.root, 0740); err != nil {
		panic(err)
	}

	syscall.Umask(oldMask)
}

func readPid() int {
	data, err := ioutil.ReadFile(files.pid)
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
	if err := ioutil.WriteFile(files.pid, data, trakxPerms); err != nil {
		panic(err)
	}
}

func clearPid() {
	pidFile, err := os.OpenFile(files.pid, os.O_CREATE|os.O_RDWR, trakxPerms)
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
	help += fmt.Sprintf("  %-12s Wipes trakx pid file\n", "wipe")
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
	logFile, err := os.OpenFile(files.log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, trakxPerms)
	if err != nil {
		panic(err)
	}
	defer logFile.Close() // TODO test
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
	files.init()

	if len(os.Args) < 2 {
		printHelp("")
		return
	}

	switch os.Args[1] {
	case "wipe":
		fmt.Println("wiping...")
		clearPid()
		fmt.Println("wiped...")
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

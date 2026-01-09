package main

import (
	"os"
	"testing"
)

const (
	testProcessID = 1
	testFilePath  = "./tmp.pid"
)

func cleanFile(t *testing.T) {
	if err := os.Remove(testFilePath); err != nil {
		if !os.IsNotExist(err) {
			t.Error("failed to remove test process id file", err)
		}
	}
}

func TestProcessIDWrite(t *testing.T) {
	defer cleanFile(t)

	pidFile := newProcessIDFile(testFilePath)
	if err := pidFile.Write(testProcessID); err != nil {
		t.Error("failed to write process id file:", err)
	}
}

func TestProcessIDRead(t *testing.T) {
	defer cleanFile(t)

	pidFile := newProcessIDFile(testFilePath)
	if err := pidFile.Write(testProcessID); err != nil {
		t.Error("failed to write process id file:", err)
	}

	processid, err := pidFile.Read()
	if err != nil {
		t.Error("failed to read process id file:", err)
	}
	if processid != testProcessID {
		t.Errorf("process id = %v; want %v", processid, testProcessID)
	}
}

func TestProcessIDClear(t *testing.T) {
	defer cleanFile(t)

	pidFile := newProcessIDFile(testFilePath)
	if err := pidFile.Write(testProcessID); err != nil {
		t.Error("failed to write process id file:", err)
	}

	if err := pidFile.Clear(); err != nil {
		t.Error("failed to clear process id file:", err)
	}

	if _, err := pidFile.Read(); err != errFileEmpty {
		t.Errorf("error = %v; want %v", err, errFileEmpty)
	}
}

func TestProcessIDProcess(t *testing.T) {
	defer cleanFile(t)

	pidFile := newProcessIDFile(testFilePath)
	if err := pidFile.Write(os.Getpid()); err != nil {
		t.Error("failed to write process id file:", err)
	}

	process, err := pidFile.Process()
	if err != nil {
		t.Error("failed to create process from process id file:", err)
	}
	if process.Pid != os.Getpid() {
		t.Errorf("process pid = %v; want %v", process.Pid, os.Getpid())
	}
}

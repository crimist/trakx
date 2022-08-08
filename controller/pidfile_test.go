package controller

import (
	"os"
	"testing"
)

const (
	testProcessID = 1
	testFilePath  = "./tmp.pid"
)

func CleanFile(t *testing.T) {
	if err := os.Remove(testFilePath); err != nil {
		t.Error("failed to remove test process id file", err)
	}
}

func TestProcessIDWrite(t *testing.T) {
	defer CleanFile(t)

	pidFile := NewProcessIDFile(testFilePath)
	err := pidFile.Write(testProcessID)
	if err != nil {
		t.Error("failed to write process id file:", err)
	}
}

func TestProcessIDRead(t *testing.T) {
	defer CleanFile(t)

	pidFile := NewProcessIDFile(testFilePath)
	err := pidFile.Write(testProcessID)
	if err != nil {
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
	defer CleanFile(t)

	pidFile := NewProcessIDFile(testFilePath)
	err := pidFile.Write(testProcessID)
	if err != nil {
		t.Error("failed to write process id file:", err)
	}

	if err := pidFile.Clear(); err != nil {
		t.Error("failed to clear process id file:", err)
	}

	if _, err := pidFile.Read(); err != ErrFileEmpty {
		t.Errorf("error = %v; want %v", err, ErrFileEmpty)
	}
}

func TestProcessIDProcess(t *testing.T) {
	defer CleanFile(t)

	pidFile := NewProcessIDFile(testFilePath)
	err := pidFile.Write(os.Getpid())
	if err != nil {
		t.Error("failed to write process id file:", err)
	}

	process, err := pidFile.Process()
	if err != nil {
		t.Error("failed to create process from proccess id file:", err)
	}
	if process.Pid != os.Getpid() {
		t.Errorf("process pid = %v; want %v", process.Pid, os.Getpid())
	}
}

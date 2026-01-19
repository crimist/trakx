package pidfile

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

func TestWrite(t *testing.T) {
	defer cleanFile(t)

	pidFile := New(testFilePath)
	if err := pidFile.Write(testProcessID); err != nil {
		t.Error("failed to write process id file:", err)
	}
}

func TestRead(t *testing.T) {
	defer cleanFile(t)

	pidFile := New(testFilePath)
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

func TestClear(t *testing.T) {
	defer cleanFile(t)

	pidFile := New(testFilePath)
	if err := pidFile.Write(testProcessID); err != nil {
		t.Error("failed to write process id file:", err)
	}

	if err := pidFile.Clear(); err != nil {
		t.Error("failed to clear process id file:", err)
	}

	if _, err := pidFile.Read(); err != ErrFileEmpty {
		t.Errorf("error = %v; want %v", err, ErrFileEmpty)
	}
}

func TestProcess(t *testing.T) {
	defer cleanFile(t)

	pidFile := New(testFilePath)
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

func TestIsAlive(t *testing.T) {
	// Current process should be alive
	if !IsAlive(os.Getpid()) {
		t.Error("current process should be alive")
	}

	// Invalid PIDs should not be alive
	if IsAlive(0) {
		t.Error("PID 0 should not be alive")
	}
	if IsAlive(-1) {
		t.Error("PID -1 should not be alive")
	}
}

func TestFileIsAlive(t *testing.T) {
	defer cleanFile(t)

	pidFile := New(testFilePath)
	if err := pidFile.Write(os.Getpid()); err != nil {
		t.Error("failed to write process id file:", err)
	}

	alive, err := pidFile.IsAlive()
	if err != nil {
		t.Error("failed to check if process is alive:", err)
	}
	if !alive {
		t.Error("current process should be alive")
	}
}

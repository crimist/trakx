# Consolidate Process ID Reading Logic - Implementation Plan

## Executive Summary

**Objective**: Consolidate duplicated process ID reading logic across the codebase into a shared `internal/pidfile` package.

**Verdict**: YES, consolidation is definitely needed. ~60+ lines of duplicate code found across 2 main locations with 3 additional duplicate liveness checks.

**Backwards Compatibility**: User explicitly stated NO NEED TO WORRY ABOUT BACKWARDS COMPATIBILITY.

## Problem Statement

Process ID reading from cache/files is duplicated across multiple locations:

### 1. Main Duplication: PID File Reading

**Location A: `cli/pidfile.go` (lines 28-42)**
- Well-structured `processIDFile` type
- Methods: `Read()`, `Write()`, `Clear()`, `Process()`
- Proper error constants: `errFileEmpty`, `errParseFailed`
- Test coverage in `cli/pidfile_test.go`
- **PROBLEM**: This is in package main, cannot be imported by other packages

**Location B: `backup/manager.go` (lines 182-202)**
- Method `readProcessID()` duplicates identical PID reading logic
- Includes debug logging with zap
- Returns -1 on error
- **This is a direct duplicate of cli/pidfile.go Read() method**

### 2. Additional Duplication: Process Liveness Checking

Duplicated in 3 locations using `syscall.Kill(processid, syscall.Signal(0))`:

1. **`backup/manager.go` (line 210)** - `isProcessAlive()` method
   - **Best implementation**: Handles EPERM correctly with `err == nil || errors.Is(err, syscall.EPERM)`
   - Includes debug logging

2. **`cli/daemon.go` (line 94)** - Inline in Stop() method
   - Used in a loop waiting for process death
   - Does NOT handle EPERM

3. **`cli/daemon.go` (line 116)** - Inline in Status() method
   - Does NOT handle EPERM

4. **`cli/cmd_pid.go` (line 89)** - Inline in pid alive command
   - Does NOT handle EPERM

## Why Current Structure Doesn't Work

**Package Import Problem:**
- `cli` is package main → cannot be imported
- `backup` package is imported by `cli` → cannot import `cli` (would create cycle)
- `daemon` package also imports `backup`

**Current dependency chain:**
```
cli (main) → imports backup
daemon → imports backup
backup → cannot import cli (it's package main)
```

**Solution:** Create `internal/pidfile` package that can be imported by all packages.

## Detailed Code Analysis

### Current Implementation: cli/pidfile.go

```go
package main

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	pidFilePermissions = 0644
	processIDFailed    = -1
)

var (
	errFileEmpty   = errors.New("process id file is empty")
	errParseFailed = errors.New("failed to parse process id file")
)

type processIDFile struct {
	path string
}

func newProcessIDFile(path string) *processIDFile {
	return &processIDFile{path: path}
}

func (pidFile *processIDFile) Read() (int, error) {
	contents, err := os.ReadFile(pidFile.path)
	if err != nil {
		return processIDFailed, err
	}
	if len(contents) == 0 {
		return processIDFailed, errFileEmpty
	}

	processID, err := strconv.Atoi(string(contents))
	if err != nil {
		return processIDFailed, errParseFailed
	}

	return processID, nil
}

func (pidFile *processIDFile) Write(processID int) error {
	return os.WriteFile(pidFile.path, []byte(strconv.Itoa(processID)), pidFilePermissions)
}

func (pidFile *processIDFile) Clear() error {
	return os.Truncate(pidFile.path, 0)
}

func (pidFile *processIDFile) Process() (*os.Process, error) {
	processid, err := pidFile.Read()
	if err != nil {
		return nil, err
	}

	process, err := os.FindProcess(processid)
	if err != nil {
		return nil, err
	}
	return process, nil
}
```

### Current Implementation: backup/manager.go (DUPLICATE)

```go
// Lines 182-202
// readProcessID reads the daemon PID from file.
func (m *Manager) readProcessID() (int, error) {
	if m.pidFilePath == "" {
		return -1, errors.New("PID file path not configured")
	}

	zap.L().Debug("Reading process ID file", zap.String("path", m.pidFilePath))
	contents, err := os.ReadFile(m.pidFilePath)
	if err != nil {
		return -1, err
	}
	if len(contents) == 0 {
		return -1, errors.New("PID file is empty")
	}

	processID, err := strconv.Atoi(string(contents))
	if err != nil {
		return -1, errors.New("failed to parse PID file")
	}

	return processID, nil
}

// Lines 204-212
// isProcessAlive checks if a process is running.
func (m *Manager) isProcessAlive(processID int) bool {
	if processID <= 0 {
		return false
	}
	zap.L().Debug("Checking process liveness", zap.Int("pid", processID))
	err := syscall.Kill(processID, syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)  // <-- BEST IMPLEMENTATION
}
```

### Current Usage: backup/manager.go selectExportSource()

```go
// Line 117
processID, err := m.readProcessID()
if err == nil && m.isProcessAlive(processID) {
	// Use socket communication with daemon
}
```

### Current Usage: cli/daemon.go

```go
// Line 31 - daemonController struct
type daemonController struct {
	processIDFile *processIDFile  // <-- Uses cli/pidfile.go type
	// ...
}

// Line 75 - constructor
func newDaemonController(conf config.Configer, out io.Writer) *daemonController {
	return &daemonController{
		processIDFile: newProcessIDFile(conf.PIDPath()),
		// ...
	}
}

// Line 94 - Stop() method - DUPLICATE LIVENESS CHECK
for err == nil && i < 100 {
	time.Sleep(100 * time.Millisecond)
	fmt.Fprint(out, ".")
	i++
	err = syscall.Kill(processid, syscall.Signal(0))  // <-- DUPLICATE
}

// Line 116 - Status() method - DUPLICATE LIVENESS CHECK
if err := syscall.Kill(processid, syscall.Signal(0)); err == nil {  // <-- DUPLICATE
	processAlive = true
}
```

### Current Usage: cli/cmd_pid.go

```go
// Line 35, 60, 84 - Multiple places create processIDFile
pidFile := newProcessIDFile(conf.PIDPath())

// Line 89 - pid alive command - DUPLICATE LIVENESS CHECK
if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {  // <-- DUPLICATE
	fmt.Fprintln(ctx.Stdout, "alive")
} else {
	fmt.Fprintln(ctx.Stdout, "dead")
}
```

## Solution: Create internal/pidfile Package

### New Package Structure

**File: `internal/pidfile/pidfile.go`**

```go
package pidfile

import (
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
)

const (
	PIDFilePermissions = 0644
	ProcessIDFailed    = -1
)

var (
	ErrFileEmpty   = errors.New("process id file is empty")
	ErrParseFailed = errors.New("failed to parse process id file")
)

// File represents a PID file for reading/writing process IDs
type File struct {
	path string
}

// New creates a new PID file handler
func New(path string) *File {
	return &File{path: path}
}

// Read reads the process ID from the file
func (f *File) Read() (int, error) {
	contents, err := os.ReadFile(f.path)
	if err != nil {
		return ProcessIDFailed, err
	}
	if len(contents) == 0 {
		return ProcessIDFailed, ErrFileEmpty
	}

	processID, err := strconv.Atoi(string(contents))
	if err != nil {
		return ProcessIDFailed, ErrParseFailed
	}

	return processID, nil
}

// Write writes a process ID to the file
func (f *File) Write(processID int) error {
	return os.WriteFile(f.path, []byte(strconv.Itoa(processID)), PIDFilePermissions)
}

// Clear truncates the PID file
func (f *File) Clear() error {
	return os.Truncate(f.path, 0)
}

// Process returns an os.Process for the PID in the file
func (f *File) Process() (*os.Process, error) {
	processID, err := f.Read()
	if err != nil {
		return nil, err
	}

	process, err := os.FindProcess(processID)
	if err != nil {
		return nil, err
	}
	return process, nil
}

// IsAlive checks if a process with the given PID is running
// Returns true if the process exists, false otherwise
// Handles EPERM (permission denied) as "alive" since the process exists
func IsAlive(processID int) bool {
	if processID <= 0 {
		return false
	}
	err := syscall.Kill(processID, syscall.Signal(0))
	// Process is alive if no error or EPERM (permission denied)
	// CRITICAL: This is the correct implementation from backup/manager.go
	return err == nil || errors.Is(err, syscall.EPERM)
}

// IsAlive reads the PID from the file and checks if that process is alive
func (f *File) IsAlive() (bool, error) {
	processID, err := f.Read()
	if err != nil {
		return false, err
	}
	return IsAlive(processID), nil
}
```

**Design Decisions:**
- Export both `File` type (for file-based operations) and `IsAlive()` function (for standalone checks)
- Use EPERM handling from backup/manager.go (most correct implementation)
- Export constants/errors with proper capitalization for public API
- Keep API similar to existing cli/pidfile.go for easier migration
- Add `File.IsAlive()` convenience method

## Implementation Steps

### Step 1: Create New Package

1. **Create directory**: `internal/pidfile/`
2. **Create file**: `internal/pidfile/pidfile.go` (use code above)
3. **Create file**: `internal/pidfile/pidfile_test.go` (port from `cli/pidfile_test.go` + add IsAlive tests)

### Step 2: Update backup/manager.go

**Import changes:**
```go
import (
	// ... existing imports ...
	"github.com/crimist/trakx/internal/pidfile"
)
```

**Line 117 - Update selectExportSource():**

BEFORE:
```go
processID, err := m.readProcessID()
if err == nil && m.isProcessAlive(processID) {
```

AFTER:
```go
pidFile := pidfile.New(m.pidFilePath)
processID, err := pidFile.Read()
if err == nil && pidfile.IsAlive(processID) {
```

**DELETE entirely:**
- Lines 181-202: `readProcessID()` method
- Lines 204-212: `isProcessAlive()` method

### Step 3: Update cli/daemon.go

**Import changes:**
```go
import (
	// ... existing imports ...
	"github.com/crimist/trakx/internal/pidfile"
)
```

**Type change in daemonController struct:**

BEFORE:
```go
type daemonController struct {
	processIDFile *processIDFile
	// ...
}
```

AFTER:
```go
type daemonController struct {
	processIDFile *pidfile.File
	// ...
}
```

**Line 75 - Update constructor:**

BEFORE:
```go
processIDFile: newProcessIDFile(conf.PIDPath()),
```

AFTER:
```go
processIDFile: pidfile.New(conf.PIDPath()),
```

**Line 94 - Replace syscall.Kill with pidfile.IsAlive():**

BEFORE:
```go
for err == nil && i < 100 {
	time.Sleep(100 * time.Millisecond)
	fmt.Fprint(out, ".")
	i++
	err = syscall.Kill(processid, syscall.Signal(0))
}
```

AFTER:
```go
for pidfile.IsAlive(processid) && i < 100 {
	time.Sleep(100 * time.Millisecond)
	fmt.Fprint(out, ".")
	i++
}
err = nil
if i < 100 {
	err = errors.New("process not alive")
}
```

**Line 116 - Replace syscall.Kill with pidfile.IsAlive():**

BEFORE:
```go
if err := syscall.Kill(processid, syscall.Signal(0)); err == nil {
	processAlive = true
}
```

AFTER:
```go
if pidfile.IsAlive(processid) {
	processAlive = true
}
```

### Step 4: Update cli/cmd_pid.go

**Import changes:**
```go
import (
	// ... existing imports ...
	"github.com/crimist/trakx/internal/pidfile"
)
```

**Lines 35, 60, 84 - Update all newProcessIDFile calls:**

BEFORE:
```go
pidFile := newProcessIDFile(conf.PIDPath())
```

AFTER:
```go
pidFile := pidfile.New(conf.PIDPath())
```

**Lines 89-96 - Replace syscall.Kill with pidfile.IsAlive():**

BEFORE:
```go
if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {
	fmt.Fprintln(ctx.Stdout, "alive")
} else {
	fmt.Fprintln(ctx.Stdout, "dead")
}
```

AFTER:
```go
if pidfile.IsAlive(pid) {
	fmt.Fprintln(ctx.Stdout, "alive")
} else {
	fmt.Fprintln(ctx.Stdout, "dead")
}
```

### Step 5: Delete Old Files

Delete these files entirely:
- `cli/pidfile.go`
- `cli/pidfile_test.go`

### Step 6: Verification

Run these commands to verify:

```bash
# Run all tests
go test ./...

# Build the project
go build ./...

# Specific package tests
go test ./internal/pidfile
go test ./backup
go test ./cli
```

## Critical Context for Implementation

### EPERM Handling is CRITICAL

The `backup/manager.go` implementation correctly handles EPERM:

```go
err := syscall.Kill(processID, syscall.Signal(0))
return err == nil || errors.Is(err, syscall.EPERM)
```

**Why this matters:**
- EPERM = "operation not permitted"
- This means the process EXISTS but we don't have permission to signal it
- This should be treated as "alive" because the process is running
- The cli implementations do NOT handle this, which is a bug

### Error String Checking (Line 100 cli/daemon.go)

```go
if err != nil && err.Error() != "no such process" {
	return errors.Wrap(err, "failed to kill trakx process id")
}
```

This is brittle string matching but doesn't affect the consolidation. The logic will continue to work after the refactoring.

### ProcessIDFailed Sentinel Value

The constant `-1` is used throughout as a sentinel value for "failed to read process ID". This is preserved in the new package as `ProcessIDFailed = -1`.

### Git Status Context

Current branch: dev
Main branch: master

Modified files relevant to this task:
- `M daemon/run.go` - May need verification
- `M daemon/backup_test.go` - May need verification
- `M storage/database.go` - Not relevant
- Several deleted files in tracker/udp - Not relevant

## Testing Strategy

1. **Port existing tests** from `cli/pidfile_test.go` to `internal/pidfile/pidfile_test.go`
2. **Add new tests** for `IsAlive()` function:
   - Test with valid PID (current process: `os.Getpid()`)
   - Test with invalid PID (0, -1, non-existent)
   - Test with PID 1 (init process, may trigger EPERM on some systems)
3. **Integration tests**: Verify backup/manager.go still works correctly
4. **CLI tests**: Verify daemon commands still work

## Potential Edge Cases

1. **PID file doesn't exist** - Read() returns error from os.ReadFile
2. **PID file is empty** - Read() returns ErrFileEmpty
3. **PID file contains non-numeric data** - Read() returns ErrParseFailed
4. **Process exists but owned by different user** - IsAlive() correctly handles EPERM
5. **PID is 0 or negative** - IsAlive() returns false
6. **PID file path not configured** (backup/manager.go line 183) - This check will be lost, but it's redundant since os.ReadFile will fail anyway

## Benefits of This Refactoring

1. ✅ **Eliminates ~60+ lines of duplicate code**
2. ✅ **Single source of truth** for all PID operations
3. ✅ **Correct EPERM handling** everywhere (fixes bug in cli)
4. ✅ **Reusable** across all packages (cli, backup, daemon)
5. ✅ **Better testability** - can test in isolation
6. ✅ **No import cycles** - internal/ can be imported by all packages
7. ✅ **Follows Go conventions** - internal/ package pattern is idiomatic
8. ✅ **Maintains API compatibility** - minimal changes to calling code

## Files Summary

### CREATE
- `internal/pidfile/pidfile.go` - 80 lines
- `internal/pidfile/pidfile_test.go` - ~50 lines

### DELETE
- `cli/pidfile.go` - 65 lines
- `cli/pidfile_test.go` - ~40 lines

### MODIFY
- `backup/manager.go` - Remove 32 lines (methods), modify 3 lines (import + usage)
- `cli/daemon.go` - Modify ~10 lines (import, type, constructor, 2 liveness checks)
- `cli/cmd_pid.go` - Modify ~8 lines (import, 3 constructors, 1 liveness check)

**Net change: -55 lines of code, +1 reusable package**

## Questions to Consider (Already Resolved)

1. **Should we move to internal/ or keep in cli?** → RESOLVED: Must use internal/ because cli is package main
2. **Which EPERM handling to use?** → RESOLVED: Use backup/manager.go version (correct)
3. **Should we consolidate liveness checking?** → RESOLVED: Yes, add IsAlive() function
4. **Backwards compatibility?** → RESOLVED: User said don't worry about it
5. **Import cycles?** → RESOLVED: internal/ solves this

## Ready to Implement

This plan is complete and ready for implementation. All code examples, line numbers, and context have been provided. A new Claude instance can follow this plan step-by-step to complete the consolidation.

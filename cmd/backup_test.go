package cmd

import (
	"bytes"
	"io"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/storage/database"
)

func TestStreamSnapshotToSocket(t *testing.T) {
	db, err := database.NewDatabase(database.Config{
		InitalSize:         1,
		PersistanceAddress: "",
		Collector:          stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}

	var hash storage.Hash
	var peerID storage.PeerID
	copy(hash[:], bytes.Repeat([]byte{1}, len(hash)))
	copy(peerID[:], bytes.Repeat([]byte{2}, len(peerID)))
	db.PeerAdd(hash, peerID, netip.MustParseAddr("127.0.0.1"), 1234, true)

	socketPath := backupSocketPath(os.Getpid())
	if err := removeSocketPath(socketPath); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	defer removeSocketPath(socketPath)

	if unixListener, ok := listener.(*net.UnixListener); ok {
		unixListener.SetDeadline(time.Now().Add(backupAcceptTimeout))
	}

	dataCh := make(chan []byte, 1)
	errCh := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			return
		}
		defer conn.Close()

		data, err := io.ReadAll(conn)
		if err != nil {
			errCh <- err
			return
		}
		dataCh <- data
	}()

	if err := streamSnapshotToSocket(db); err != nil {
		t.Fatalf("streamSnapshotToSocket failed: %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("listener error: %v", err)
	case data := <-dataCh:
		restored, err := database.NewDatabase(database.Config{
			InitalSize:         1,
			PersistanceAddress: "",
			Collector:          stats.NewCollectors(false, false, 0),
		})
		if err != nil {
			t.Fatal("Failed to create database")
		}
		if err := restored.Restore(bytes.NewReader(data)); err != nil {
			t.Fatalf("Restore failed: %v", err)
		}
		if restored.Torrents() != 1 {
			t.Fatalf("torrents = %d, want 1", restored.Torrents())
		}
	case <-time.After(backupAcceptTimeout + time.Second):
		t.Fatal("timed out waiting for snapshot")
	}
}

func TestExportBackupFallbackFile(t *testing.T) {
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "db")
	expected := []byte("backup-bytes")
	if err := os.WriteFile(backupPath, expected, backupFilePermissions); err != nil {
		t.Fatal(err)
	}

	conf := &config.Configuration{}
	conf.Cache = tmpDir
	conf.DB.Backup.Path = backupPath

	var buf bytes.Buffer
	if err := ExportBackup(conf, &buf); err != nil {
		t.Fatalf("ExportBackup failed: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Fatalf("backup bytes = %q, want %q", buf.Bytes(), expected)
	}
}

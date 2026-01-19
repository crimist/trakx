package daemon

import (
	"bytes"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crimist/trakx/backup"
	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/internal/pidfile"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/storage/database"
	"github.com/crimist/trakx/tracker/udp/connections"
)

func TestWriteCombinedSnapshot(t *testing.T) {
	db, err := database.NewDatabase(database.Config{
		InitalSize: 1,
		Collector:  stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}

	var hash storage.Hash
	var peerID storage.PeerID
	copy(hash[:], bytes.Repeat([]byte{1}, len(hash)))
	copy(peerID[:], bytes.Repeat([]byte{2}, len(peerID)))
	db.PeerAdd(hash, peerID, netip.MustParseAddr("127.0.0.1"), 1234, true)

	connDB := connections.NewConnections(1, time.Minute, 0)
	udpAddr := netip.MustParseAddrPort("1.1.1.1:1234")
	connID, err := connDB.Create(udpAddr)
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}

	// Write combined snapshot to buffer
	var buf bytes.Buffer
	if err := backup.WriteCombined(&buf, db, connDB); err != nil {
		t.Fatalf("WriteCombined failed: %v", err)
	}

	// Read it back
	connSnapshot, dbReader, err := backup.RestoreCombined(&buf)
	if err != nil {
		t.Fatalf("RestoreCombined failed: %v", err)
	}

	restored, err := database.NewDatabase(database.Config{
		InitalSize: 1,
		Collector:  stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}
	if err := restored.Restore(dbReader); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if restored.Torrents() != 1 {
		t.Fatalf("torrents = %d, want 1", restored.Torrents())
	}

	restoredConnections := connections.NewConnections(1, time.Minute, 0)
	if err := restoredConnections.Unmarshal(connSnapshot); err != nil {
		t.Fatalf("Failed to restore connections: %v", err)
	}
	if !restoredConnections.Validate(udpAddr, connID) {
		t.Fatalf("restored connections missing expected entry")
	}
}

func TestExportBackupFallbackFile(t *testing.T) {
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "db")
	expected := []byte("backup-bytes")
	if err := os.WriteFile(backupPath, expected, 0640); err != nil {
		t.Fatal(err)
	}

	conf := &config.Configuration{}
	conf.Cache = tmpDir
	conf.DB.Backup.Path = backupPath

	mgr := backup.NewManager(backup.Config{
		BackupFilePath: conf.DB.Backup.Path,
		PIDFile:        pidfile.New(conf.PIDPath()),
		CacheDir:       conf.Cache,
	})

	var buf bytes.Buffer
	dest := backup.NewStreamDestination(&buf, "test")
	if err := mgr.Export(dest); err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Fatalf("backup bytes = %q, want %q", buf.Bytes(), expected)
	}
}

func TestImportBackupWritesValidData(t *testing.T) {
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "db")

	conf := &config.Configuration{}
	conf.Cache = tmpDir
	conf.DB.Backup.Path = backupPath

	mgr := backup.NewManager(backup.Config{
		BackupFilePath: conf.DB.Backup.Path,
		PIDFile:        pidfile.New(conf.PIDPath()),
		CacheDir:       conf.Cache,
	})

	// Create a valid backup
	db, err := database.NewDatabase(database.Config{
		InitalSize: 1,
		Collector:  stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}

	var buf bytes.Buffer
	if err := backup.WriteCombined(&buf, db, nil); err != nil {
		t.Fatal(err)
	}

	source := backup.NewStreamSource(&buf, "test")
	if err := mgr.Import(source); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("backup file was not created")
	}
}

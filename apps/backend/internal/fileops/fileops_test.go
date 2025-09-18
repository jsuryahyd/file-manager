package fileops

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"file-manager-backend/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestSync(t *testing.T) (*sql.DB, string, string, func()) {
	// Setup database
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	initSQL := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL,
		size INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sync_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_dir TEXT NOT NULL,
		dest_dir TEXT NOT NULL,
		started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		status TEXT NOT NULL CHECK(status IN ('running', 'completed', 'failed'))
	);

	CREATE TABLE IF NOT EXISTS synced_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sync_job_id INTEGER NOT NULL,
		file_id INTEGER NOT NULL,
		FOREIGN KEY (sync_job_id) REFERENCES sync_jobs(id),
		FOREIGN KEY (file_id) REFERENCES files(id)
	);
	`
	tmpfile, err := os.CreateTemp("", "init-*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp init.sql: %v", err)
	}

	if _, err := tmpfile.Write([]byte(initSQL)); err != nil {
		t.Fatalf("Failed to write to temp init.sql: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp init.sql: %v", err)
	}

	if err := db.Migrate(database, tmpfile.Name()); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Setup directories
	srcDir, err := os.MkdirTemp("", "src")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}

	dstDir, err := os.MkdirTemp("", "dst")
	if err != nil {
		t.Fatalf("Failed to create temp dst dir: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.RemoveAll(srcDir)
		os.RemoveAll(dstDir)
		os.Remove(tmpfile.Name())
	}

	return database, srcDir, dstDir, cleanup
}

func TestSyncUniqueFiles(t *testing.T) {
	database, srcDir, dstDir, cleanup := setupTestSync(t)
	defer cleanup()

	// Create a test file
	fileContent := []byte("hello world")
	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, fileContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// First sync
	copied, err := SyncUniqueFiles(database, srcDir, dstDir)
	if err != nil {
		t.Fatalf("SyncUniqueFiles failed: %v", err)
	}

	if len(copied) != 1 {
		t.Fatalf("Expected 1 file to be copied, got %d", len(copied))
	}

	if copied[0] != "test.txt" {
		t.Errorf("Expected 'test.txt' to be copied, got %s", copied[0])
	}

	// Second sync (should not copy anything)
	copied, err = SyncUniqueFiles(database, srcDir, dstDir)
	if err != nil {
		t.Fatalf("SyncUniqueFiles failed on second run: %v", err)
	}

	if len(copied) != 0 {
		t.Fatalf("Expected 0 files to be copied on second run, got %d", len(copied))
	}
}

package fileops

import (
	"database/sql"
	"path/filepath"
	"testing"

	"file-manager-backend/internal/db"
	"github.com/spf13/afero"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestSync(t *testing.T) (*sql.DB, string, string, int64, func()) {
	// Setup filesystems
	AppFs = afero.NewMemMapFs()
	db.AppFs = AppFs
	srcDir := "/src"
	dstDir := "/dst"
	AppFs.Mkdir(srcDir, 0755)
	AppFs.Mkdir(dstDir, 0755)

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

	CREATE TABLE IF NOT EXISTS sync_pairs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_dir TEXT NOT NULL,
		dest_dir TEXT NOT NULL,
		UNIQUE(source_dir, dest_dir)
	);

	CREATE TABLE IF NOT EXISTS sync_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sync_pair_id INTEGER NOT NULL,
		started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		status TEXT NOT NULL CHECK(status IN ('running', 'completed', 'failed')),
		FOREIGN KEY (sync_pair_id) REFERENCES sync_pairs(id)
	);

	CREATE TABLE IF NOT EXISTS synced_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sync_job_id INTEGER NOT NULL,
		file_id INTEGER NOT NULL,
		FOREIGN KEY (sync_job_id) REFERENCES sync_jobs(id),
		FOREIGN KEY (file_id) REFERENCES files(id)
	);
	`
	sqlPath := "/init.sql"
	afero.WriteFile(AppFs, sqlPath, []byte(initSQL), 0644)

	if err := db.Migrate(database, sqlPath); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create a sync pair for the test
	syncPairID, err := db.CreateSyncPair(database, srcDir, dstDir)
	if err != nil {
		t.Fatalf("Failed to create sync pair: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, srcDir, dstDir, syncPairID, cleanup
}

func TestSyncUniqueFiles(t *testing.T) {
	database, srcDir, dstDir, syncPairID, cleanup := setupTestSync(t)
	defer cleanup()

	// Create a test file
	fileContent := []byte("hello world")
	testFile := filepath.Join(srcDir, "test.txt")
	afero.WriteFile(AppFs, testFile, fileContent, 0644)

	// First sync
	copied, err := SyncUniqueFiles(database, srcDir, dstDir, syncPairID)
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
	copied, err = SyncUniqueFiles(database, srcDir, dstDir, syncPairID)
	if err != nil {
		t.Fatalf("SyncUniqueFiles failed on second run: %v", err)
	}

	if len(copied) != 0 {
		t.Fatalf("Expected 0 files to be copied on second run, got %d", len(copied))
	}
}

func TestListFiles(t *testing.T) {
	AppFs = afero.NewMemMapFs()
	dir := "/test"
	AppFs.Mkdir(dir, 0755)
	afero.WriteFile(AppFs, "/test/file1.txt", []byte("file1"), 0644)
	afero.WriteFile(AppFs, "/test/file2.txt", []byte("file2"), 0644)
	AppFs.Mkdir("/test/subdir", 0755)

	files, err := ListFiles(dir)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(files))
	}
}

package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create a temporary init.sql file for migration
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
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(initSQL)); err != nil {
		t.Fatalf("Failed to write to temp init.sql: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp init.sql: %v", err)
	}

	if err := Migrate(db, tmpfile.Name()); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestCreateAndGetFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	path := "/tmp/testfile.txt"
	hash := "somehash"
	size := int64(123)

	fileID, err := CreateFile(db, path, hash, size)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	if fileID == 0 {
		t.Fatal("Expected a non-zero file ID")
	}

	file, err := GetFileByPath(db, path)
	if err != nil {
		t.Fatalf("GetFileByPath failed: %v", err)
	}

	if file.Path != path {
		t.Errorf("Expected path %s, got %s", path, file.Path)
	}

	if file.Hash != hash {
		t.Errorf("Expected hash %s, got %s", hash, file.Hash)
	}

	if file.Size != size {
		t.Errorf("Expected size %d, got %d", size, file.Size)
	}
}

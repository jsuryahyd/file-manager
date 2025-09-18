package db

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// File represents a file in the database.
type File struct {
	ID         int64
	Path       string
	Hash       string
	Size       int64
	CreatedAt  string
	ModifiedAt string
}

// InitDB initializes the SQLite database and returns the connection.
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate runs the database initialization SQL script.
func Migrate(db *sql.DB, sqlPath string) error {
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(content))
	return err
}

// CreateFile adds a new file to the database.
func CreateFile(db *sql.DB, path, hash string, size int64) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO files(path, hash, size) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(path, hash, size)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// GetFileByPath retrieves a file from the database by its path.
func GetFileByPath(db *sql.DB, path string) (*File, error) {
	row := db.QueryRow("SELECT id, path, hash, size, created_at, modified_at FROM files WHERE path = ?", path)

	file := &File{}
	err := row.Scan(&file.ID, &file.Path, &file.Hash, &file.Size, &file.CreatedAt, &file.ModifiedAt)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// CreateSyncJob adds a new sync job to the database.
func CreateSyncJob(db *sql.DB, sourceDir, destDir string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO sync_jobs(source_dir, dest_dir, status) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(sourceDir, destDir, "running")
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// UpdateSyncJobStatus updates the status of a sync job.
func UpdateSyncJobStatus(db *sql.DB, jobID int64, status string) error {
	_, err := db.Exec("UPDATE sync_jobs SET status = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?", status, jobID)
	return err
}

// CreateSyncedFile links a file to a sync job.
func CreateSyncedFile(db *sql.DB, jobID, fileID int64) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO synced_files(sync_job_id, file_id) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(jobID, fileID)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

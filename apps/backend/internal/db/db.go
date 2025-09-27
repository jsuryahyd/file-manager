package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/afero"
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

// SyncPair represents a source-destination pair for synchronization.
type SyncPair struct {
	ID        int64
	SourceDir string
	DestDir   string
}

// SyncJob represents a single sync job with its associated pair info.
type SyncJob struct {
	ID          int64  `json:"id"`
	SyncPairID  int64  `json:"syncPairId"`
	Status      string `json:"status"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
	SourceDir   string `json:"sourceDir"`
	DestDir     string `json:"destDir"`
	Misc        string `json:"misc"`
}

// InitDB initializes the SQLite database and returns the connection.
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate runs the database initialization and migration scripts.
func Migrate(db *sql.DB, initSQLPath string, migrationsPath string) error {
	// Run initial schema
	content, err := afero.Afero{Fs: AppFs}.ReadFile(initSQLPath)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(content))
	if err != nil {
		return err
	}

	// Create migrations table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT NOT NULL PRIMARY KEY);`)
	if err != nil {
		return err
	}

	// Get applied migrations
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return err
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		appliedMigrations[version] = true
	}

	// Run migrations
	migrations, err := afero.ReadDir(AppFs, migrationsPath)
	if err != nil {
		// Migrations directory might not exist, which is fine
		return nil
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name() < migrations[j].Name()
	})

	for _, migration := range migrations {
		if filepath.Ext(migration.Name()) == ".sql" {
			if !appliedMigrations[migration.Name()] {
				migrationPath := filepath.Join(migrationsPath, migration.Name())
				content, err := afero.Afero{Fs: AppFs}.ReadFile(migrationPath)
				if err != nil {
					return err
				}
				_, err = db.Exec(string(content))
				if err != nil {
					return err
				}

				_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Name())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
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

// GetSyncPair retrieves a sync pair by its source and destination directories.
func GetSyncPair(db *sql.DB, sourceDir, destDir string) (*SyncPair, error) {
	row := db.QueryRow("SELECT id, source_dir, dest_dir FROM sync_pairs WHERE source_dir = ? AND dest_dir = ?", sourceDir, destDir)

	pair := &SyncPair{}
	err := row.Scan(&pair.ID, &pair.SourceDir, &pair.DestDir)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

// CreateSyncPair creates a new sync pair.
func CreateSyncPair(db *sql.DB, sourceDir, destDir string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO sync_pairs(source_dir, dest_dir) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(sourceDir, destDir)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CreateSyncJob adds a new sync job to the database.
func CreateSyncJob(db *sql.DB, syncPairID int64) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO sync_jobs(sync_pair_id, status) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(syncPairID, "running")
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// UpdateSyncJobStatus updates the status of a sync job.
func UpdateSyncJobStatus(db *sql.DB, jobID int64, status string, errorMsg string) error {
	var err error
	if status == "failed" {
		misc := fmt.Sprintf(`{"error": "%s"}`, errorMsg)
		_, err = db.Exec("UPDATE sync_jobs SET status = ?, completed_at = CURRENT_TIMESTAMP, misc = ? WHERE id = ?", status, misc, jobID)
	} else {
		_, err = db.Exec("UPDATE sync_jobs SET status = ?, completed_at = CURRENT_TIMESTAMP, misc = NULL WHERE id = ?", status, jobID)
	}
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

// GetSyncJobs retrieves the most recent sync jobs from the database.
func GetSyncJobs(db *sql.DB, limit int) ([]SyncJob, error) {
	rows, err := db.Query(`
		SELECT
			sj.id,
			sj.sync_pair_id,
			sj.status,
			sj.started_at,
			sj.completed_at,
			sp.source_dir,
			sp.dest_dir,
			sj.misc
		FROM sync_jobs sj
		JOIN sync_pairs sp ON sj.sync_pair_id = sp.id
		ORDER BY sj.started_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []SyncJob
	for rows.Next() {
		var job SyncJob
		var completedAt sql.NullString
		var misc sql.NullString
		if err := rows.Scan(
			&job.ID,
			&job.SyncPairID,
			&job.Status,
			&job.StartedAt,
			&completedAt,
			&job.SourceDir,
			&job.DestDir,
			&misc,
		); err != nil {
			return nil, err
		}
		job.CompletedAt = completedAt.String
		job.Misc = misc.String
		jobs = append(jobs, job)
	}

	return jobs, nil
}

-- Initial schema for the file manager database

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

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
CREATE INDEX IF NOT EXISTS idx_files_hash ON files(hash);
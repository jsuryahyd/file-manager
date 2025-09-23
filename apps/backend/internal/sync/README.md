# Sync Module

This module handles file synchronization between two directories. It is designed to be robust, extensible, and performant.

## Domain-Driven Design (DDD)

The sync module follows DDD principles. The core domain is centered around the `SyncJob` aggregate.

### Aggregates

- **`SyncJob`**: Represents a single synchronization task. It encapsulates all the rules and logic for a sync operation.

### Entities

- **`SyncPair`**: Represents the source and destination directories for a sync operation.
- **`File`**: Represents a file being synced.

### Value Objects

- **`SyncOptions`**: Represents the options for a sync job (e.g., recursive, skip patterns, peek mode).
- **`SyncResult`**: Represents the result of a sync job.

## Features

- **Recursive Sync**: Synchronize files and directories recursively.
- **Skip Patterns**: Exclude files and directories based on glob patterns (e.g., `node_modules`, `*.tmp`).
- **Peek Mode (Dry Run)**: Preview the changes without actually modifying any files.
- **Conflict Resolution**: Handle cases where a file already exists at the destination.
- **Error Handling**: Robust error handling and reporting.

## API

The sync module will expose a simple API:

```go
package sync

// Options defines the options for a sync job.
type Options struct {
    Recursive    bool
    SkipPatterns []string
    PeekMode     bool
    Overwrite    bool
}

// Result holds the outcome of a sync operation.
type Result struct {
    FilesCopied  []string
    FilesSkipped []string
    Errors       []error
}

// NewSyncJob creates a new sync job.
func NewSyncJob(db *sql.DB, srcDir, dstDir string, opts Options) (*Job, error) {
    // ...
}

// Run executes the sync job.
func (j *Job) Run() (*Result, error) {
    // ...
}
```

## Future Enhancements

- **Real-time Sync**: Watch for file system changes and sync them in real-time.
- **Cloud Storage**: Sync with cloud storage providers (e.g., S3, Google Drive).
- **Performance Optimizations**: Use concurrent operations to speed up hashing and copying.

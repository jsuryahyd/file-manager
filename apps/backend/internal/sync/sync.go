package sync

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	// "strings"
	"time"

	"file-manager-backend/internal/db"
	"file-manager-backend/internal/fileops"

	"github.com/spf13/afero"
)

// Options defines the options for a sync job.
type Options struct {
	Recursive       bool
	SkipPatterns    []string
	PeekMode        bool
	Overwrite       bool
	CheckDuplicates bool
}

// Result holds the outcome of a sync operation.
type Result struct {
	FilesCopied  []string
	FilesSkipped []string
	Errors       []error
}

// Job represents a single synchronization task.
type Job struct {
	db         *sql.DB
	srcDir     string
	dstDir     string
	opts       Options
	syncPairID int64
}

// NewSyncJob creates a new sync job.
func NewSyncJob(db *sql.DB, srcDir, dstDir string, syncPairID int64, opts Options) (*Job, error) {
	return &Job{
		db:         db,
		srcDir:     srcDir,
		dstDir:     dstDir,
		opts:       opts,
		syncPairID: syncPairID,
	}, nil
}

func splitFileName(fileName string) (string, string) {
	ext := filepath.Ext(fileName)
	base := fileName[:len(fileName)-len(ext)]
	return base, ext
}

// Run executes the sync job.
func (j *Job) Run() (*Result, error) {
	// Sanitize paths
	j.srcDir = filepath.Clean(j.srcDir)
	j.dstDir = filepath.Clean(j.dstDir)

	if j.srcDir == j.dstDir {
		return nil, errors.New("source and destination cannot be the same")
	}

	jobID, err := db.CreateSyncJob(j.db, j.syncPairID)
	if err != nil {
		return nil, err
	}

	result := &Result{}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil // continue walking
		}

		// Skip symlinks to avoid issues with Windows junctions and recursive loops.
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// TODO: Implement SkipPatterns logic

		if info.IsDir() {
			if !j.opts.Recursive && path != j.srcDir {
				return filepath.SkipDir
			}
			return nil // continue walking
		}

		srcPath := path
		relPath, err := filepath.Rel(j.srcDir, srcPath)
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}
		dstPath := filepath.Join(j.dstDir, relPath)
		finalDstPath := dstPath

		// --- Start of decision logic ---
		var shouldCopy = true
		var skipReason = ""

		dstInfo, err := fileops.AppFs.Stat(dstPath)
		if err == nil { // Destination exists
			if dstInfo.IsDir() {
				// Name conflict: a directory with the same name exists.
				// Find a new name for the file.
				base, ext := splitFileName(relPath)
				i := 1
				for {
					newRelPath := fmt.Sprintf("%s (%d)%s", base, i, ext)
					newDstPath := filepath.Join(j.dstDir, newRelPath)
					if _, err := fileops.AppFs.Stat(newDstPath); os.IsNotExist(err) {
						finalDstPath = newDstPath
						break
					}
					i++
				}
			} else {
				// It's a file, check for overwrite/duplicates
				srcHash, hashErr := fileHash(srcPath)
				if hashErr != nil {
					result.Errors = append(result.Errors, hashErr)
					return nil
				}
				dstHash, hashErr := fileHash(dstPath)
				if hashErr != nil {
					result.Errors = append(result.Errors, hashErr)
					return nil
				}

				if j.opts.CheckDuplicates && srcHash == dstHash {
					shouldCopy = false
					skipReason = "identical file already exists at destination"
				} else if !j.opts.Overwrite {
					shouldCopy = false
					skipReason = "different file exists at destination and overwrite is disabled"
				}
			}
		} else if !os.IsNotExist(err) {
			shouldCopy = false
			skipReason = fmt.Sprintf("error checking destination: %v", err)
		}
		// --- End of decision logic ---

		if !shouldCopy {
			log.Printf("Skipping '%s': %s.", relPath, skipReason)
			result.FilesSkipped = append(result.FilesSkipped, relPath)
			return nil
		}

		if j.opts.PeekMode {
			log.Printf("Peek Mode: Would copy '%s' to '%s'", relPath, finalDstPath)
			result.FilesCopied = append(result.FilesCopied, relPath)
			return nil
		}

		// Create subdirectory in destination if it doesn't exist
		dstParentDir := filepath.Dir(finalDstPath)
		if _, err := fileops.AppFs.Stat(dstParentDir); os.IsNotExist(err) {
			if err := fileops.AppFs.MkdirAll(dstParentDir, 0755); err != nil {
				result.Errors = append(result.Errors, err)
				return nil
			}
		}

		srcInfo, err := fileops.AppFs.Stat(srcPath)
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil // skip if we can't stat source
		}

		srcHash, err := fileHash(srcPath)
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil // skip if we can't hash source
		}

		err = fileops.CopyFile(srcPath, finalDstPath)
		if err != nil {
			db.UpdateSyncJobStatus(j.db, jobID, "failed")
			return err
		}

		fileID, err := db.CreateFile(j.db, srcPath, srcHash, srcInfo.Size())
		if err != nil {
			db.UpdateSyncJobStatus(j.db, jobID, "failed")
			return err
		}

		_, err = db.CreateSyncedFile(j.db, jobID, fileID)
		if err != nil {
			db.UpdateSyncJobStatus(j.db, jobID, "failed")
			return err
		}

		result.FilesCopied = append(result.FilesCopied, relPath)
		return nil
	}

	err = afero.Walk(fileops.AppFs, j.srcDir, walkFunc)
	if err != nil {
		db.UpdateSyncJobStatus(j.db, jobID, "failed")
		return nil, err
	}

	db.UpdateSyncJobStatus(j.db, jobID, "completed")
	return result, nil
}

// fileHash returns the SHA256 hash of a file.
func fileHash(path string) (string, error) {
	start := time.Now()
	f, err := fileops.AppFs.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	elapsed := time.Since(start)
	log.Printf("Hashed file %s in %s", path, elapsed)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

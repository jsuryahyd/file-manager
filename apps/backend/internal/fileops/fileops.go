package fileops

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"file-manager-backend/internal/db"
	"github.com/spf13/afero"
)

// FileEntryInfo defines the structure for file and directory metadata.
type FileEntryInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Path  string `json:"path"`
}

// SyncUniqueFiles copies only unique files from srcDir to dstDir, skipping duplicates.
func SyncUniqueFiles(database *sql.DB, srcDir, dstDir string, syncPairID int64) ([]string, error) {
	// Sanitize paths to remove trailing slashes
	srcDir = strings.TrimRight(srcDir, "/")
	dstDir = strings.TrimRight(dstDir, "/")

	if srcDir == dstDir {
		return nil, errors.New("source and destination cannot be the same")
	}

	jobID, err := db.CreateSyncJob(database, syncPairID)
	if err != nil {
		return nil, err
	}

	files, err := ListFiles(srcDir)
	if err != nil {
		return nil, err
	}

	var copied []string
	for _, file := range files {
		// TODO: Implement recursive sync with a `recursive:true` parameter.
		// For now, we only sync files in the root of the source directory.
		if file.IsDir {
			continue
		}
		srcPath := file.Path
		info, err := AppFs.Stat(srcPath)
		if err != nil {
			continue // skip directories and errors
		}

		hash, err := fileHash(srcPath)
		if err != nil {
			return copied, err
		}

		existingFile, err := db.GetFileByPath(database, srcPath)
		if err != nil && err != sql.ErrNoRows {
			return copied, err
		}

		if existingFile == nil || existingFile.Hash != hash {
			dstPath := filepath.Join(dstDir, file.Name)
			err := CopyFile(srcPath, dstPath)
			if err != nil {
				db.UpdateSyncJobStatus(database, jobID, "failed")
				return copied, err
			}

			fileID, err := db.CreateFile(database, srcPath, hash, info.Size())
			if err != nil {
				db.UpdateSyncJobStatus(database, jobID, "failed")
				return copied, err
			}

			_, err = db.CreateSyncedFile(database, jobID, fileID)
			if err != nil {
				db.UpdateSyncJobStatus(database, jobID, "failed")
				return copied, err
			}

			copied = append(copied, file.Name)
		}
	}

	db.UpdateSyncJobStatus(database, jobID, "completed")
	return copied, nil
}

// ListFiles returns a list of files and folders in the given directory.
func ListFiles(dir string) ([]FileEntryInfo, error) {
	entries, err := afero.ReadDir(AppFs, dir)
	if err != nil {
		return nil, err
	}
	var files []FileEntryInfo
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		files = append(files, FileEntryInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Path:  fullPath,
		})
	}
	return files, nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	srcFile, err := AppFs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := AppFs.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// MoveFile moves a file from src to dst.
func MoveFile(src, dst string) error {
	return AppFs.Rename(src, dst)
}

// DeleteFile deletes the specified file.
func DeleteFile(path string) error {
	return AppFs.Remove(path)
}

// fileHash returns the SHA256 hash of a file.
func fileHash(path string) (string, error) {
	f, err := AppFs.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

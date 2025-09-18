package fileops

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"file-manager-backend/internal/db"
)

// SyncUniqueFiles copies only unique files from srcDir to dstDir, skipping duplicates.
func SyncUniqueFiles(database *sql.DB, srcDir, dstDir string) ([]string, error) {
	jobID, err := db.CreateSyncJob(database, srcDir, dstDir)
	if err != nil {
		return nil, err
	}

	files, err := ListFileNames(srcDir)
	if err != nil {
		return nil, err
	}

	var copied []string
	for _, name := range files {
		srcPath := filepath.Join(srcDir, name)
		info, err := os.Stat(srcPath)
		if err != nil || info.IsDir() {
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
			dstPath := filepath.Join(dstDir, name)
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

			copied = append(copied, name)
		}
	}

	db.UpdateSyncJobStatus(database, jobID, "completed")
	return copied, nil
}

// ListFileNames returns a list of files and folders in the given directory.
func ListFileNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// MoveFile moves a file from src to dst.
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// DeleteFile deletes the specified file.
func DeleteFile(path string) error {
	return os.Remove(path)
}

// fileHash returns the SHA256 hash of a file.
func fileHash(path string) (string, error) {
	f, err := os.Open(path)
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

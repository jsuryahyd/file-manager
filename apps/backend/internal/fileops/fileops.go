package fileops

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

// SyncUniqueFiles copies only unique files from srcDir to dstDir, with options to handle existing files.
func SyncUniqueFiles(database *sql.DB, srcDir, dstDir string, syncPairID int64, checkDuplicates bool, overwriteExisting bool) ([]string, error) {
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
		db.UpdateSyncJobStatus(database, jobID, "failed")
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
		dstPath := filepath.Join(dstDir, file.Name)

		srcInfo, err := AppFs.Stat(srcPath)
		if err != nil {
			continue // skip if we can't stat source
		}

		dstInfo, err := AppFs.Stat(dstPath)
		if err == nil { // Destination exists
			if dstInfo.IsDir() {
				log.Printf("Skipping '%s': a directory with the same name exists at destination.", file.Name)
				continue // Can't replace a dir with a file
			}

			var srcHash, dstHash string
			var hashErr error

			// We need hashes to compare files
			srcHash, hashErr = fileHash(srcPath)
			if hashErr != nil {
				continue // skip if we can't hash source
			}
			dstHash, hashErr = fileHash(dstPath)
			if hashErr != nil {
				continue // skip if we can't hash destination
			}

			if checkDuplicates && srcHash == dstHash {
				log.Printf("Skipping '%s': identical file already exists at destination.", file.Name)
				// Identical file exists, and we want to prevent duplicates.
				// So, we skip copying. But we ensure it's tracked in the DB.
				existingFile, err := db.GetFileByPath(database, srcPath)
				if err != nil && err != sql.ErrNoRows {
					return copied, err
				}
				if existingFile == nil { // Not tracked yet, so let's track it.
					fileID, err := db.CreateFile(database, srcPath, srcHash, srcInfo.Size())
					if err != nil {
						db.UpdateSyncJobStatus(database, jobID, "failed")
						return copied, err
					}
					_, err = db.CreateSyncedFile(database, jobID, fileID)
					if err != nil {
						db.UpdateSyncJobStatus(database, jobID, "failed")
						return copied, err
					}
				}
				continue // Done with this file.
			}

			if !overwriteExisting {
				log.Printf("Skipping '%s': different file exists at destination and overwrite is disabled.", file.Name)
				// Destination exists, it's different (or we didn't check for duplicates),
				// and we are not allowed to overwrite.
				continue
			}
		} else if !os.IsNotExist(err) {
			log.Printf("Skipping '%s': error checking destination: %v", file.Name, err)
			// Another error occurred when stating destination file (e.g. permission denied)
			continue
		}

		// If we get here, we should copy the file.
		if err == nil { // from Stat, so file exists
			log.Printf("Copying '%s': overwriting different file at destination.", file.Name)
		} else { // Stat returned an error, presumably IsNotExist
			log.Printf("Copying '%s': file does not exist at destination.", file.Name)
		}

		srcHash, err := fileHash(srcPath)
		if err != nil {
			continue // skip if we can't hash source
		}

		err = CopyFile(srcPath, dstPath)
		if err != nil {
			db.UpdateSyncJobStatus(database, jobID, "failed")
			return copied, err
		}

		fileID, err := db.CreateFile(database, srcPath, srcHash, srcInfo.Size())
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

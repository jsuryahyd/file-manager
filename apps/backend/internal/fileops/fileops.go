package fileops

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"file-manager-backend/internal/db"
	"github.com/spf13/afero"
)

// FileEntryInfo defines the structure for file and directory metadata.
type FileEntryInfo struct {
	Name                 string `json:"name"`
	IsDir                bool   `json:"isDir"`
	Path                 string `json:"path"`
	IsPotentialDuplicate bool   `json:"isPotentialDuplicate,omitempty"`
}

// ListFiles returns a list of files and folders in the given directory.
func ListFiles(dir string, fileType string) ([]FileEntryInfo, error) {
	entries, err := afero.ReadDir(db.AppFs, dir)
	if err != nil {
		return nil, err
	}
	var files []FileEntryInfo
	for _, entry := range entries {
		isDir := entry.IsDir()
		if fileType == "dir" && !isDir {
			continue
		}
		if fileType == "file" && isDir {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		files = append(files, FileEntryInfo{
			Name:  entry.Name(),
			IsDir: isDir,
			Path:  filepath.ToSlash(fullPath),
		})
	}
	return files, nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	srcFile, err := db.AppFs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := db.AppFs.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// MoveFile moves a file from src to dst.
func MoveFile(src, dst string) error {
	return db.AppFs.Rename(src, dst)
}

// DeleteFile deletes the specified file.
func DeleteFile(path string) error {
	return db.AppFs.Remove(path)
}

// DeleteFiles deletes a list of files.
func DeleteFiles(paths []string) error {
	for _, path := range paths {
		err := DeleteFile(path)
		if err != nil {
			log.Printf("Failed to delete file %s: %v", path, err)
			// Maybe collect errors and return them
		}
	}
	return nil
}

// fileHash returns the SHA256 hash of a file.
func fileHash(path string) (string, error) {
	start := time.Now()
	f, err := db.AppFs.Open(path)
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

// FindDuplicates recursively finds duplicate files in a directory. It first groups files by size and then calculates hashes only for files with the same size to optimize performance.
// TODO: Implement other strategies like metadata and name pattern matching.
func FindDuplicates(dir string) ([][]FileEntryInfo, error) {
	filesBySize := make(map[int64][]FileEntryInfo)

	err := afero.Walk(db.AppFs, dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // continue walking
		}
		fileEntry := FileEntryInfo{
			Name:  info.Name(),
			IsDir: false,
			Path:  path,
		}
		filesBySize[info.Size()] = append(filesBySize[info.Size()], fileEntry)
		return nil
	})

	if err != nil {
		return nil, err
	}

	var duplicates [][]FileEntryInfo
	const oneHundredMB = 100 * 1024 * 1024

	for size, files := range filesBySize {
		if len(files) < 2 {
			continue // No potential duplicates in this size group
		}

		// If file size is greater than 100MB, mark as potential duplicates without hashing.
		if size > oneHundredMB {
			log.Printf("Files with size %d (> 100MB) are marked as potential duplicates without hashing.", size)
			var potentialDuplicates []FileEntryInfo
			for _, file := range files {
				file.IsPotentialDuplicate = true
				potentialDuplicates = append(potentialDuplicates, file)
			}
			duplicates = append(duplicates, potentialDuplicates)
			continue
		}

		// Potential duplicates found based on size. Now hash them.
		hashes := make(map[string][]FileEntryInfo)
		for _, file := range files {
			hash, err := fileHash(file.Path)
			if err != nil {
				log.Printf("Could not hash file %s: %v", file.Path, err)
				continue
			}
			hashes[hash] = append(hashes[hash], file)
		}

		// Add actual duplicates to the result
		for _, hashGroup := range hashes {
			if len(hashGroup) > 1 {
				duplicates = append(duplicates, hashGroup)
			}
		}
	}

	return duplicates, nil
}
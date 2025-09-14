package fileops

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SyncUniqueFiles copies only unique files from srcDir to dstDir, skipping duplicates.
func SyncUniqueFiles(srcDir, dstDir string) ([]string, error) {
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
		isDup, err := IsDuplicate(srcPath, dstDir)
		if err != nil {
			return copied, err
		}
		if !isDup {
			dstPath := filepath.Join(dstDir, name)
			err := CopyFile(srcPath, dstPath)
			if err != nil {
				return copied, err
			}
			copied = append(copied, name)
		}
	}
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
	_, err = srcFile.WriteTo(dstFile)
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

// IsDuplicate checks if a file with the same name, size, and hash exists in the destination directory.
func IsDuplicate(src, dstDir string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	dstPath := filepath.Join(dstDir, filepath.Base(src))
	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if srcInfo.Size() != dstInfo.Size() {
		return false, nil
	}
	srcHash, err := fileHash(src)
	if err != nil {
		return false, err
	}
	dstHash, err := fileHash(dstPath)
	if err != nil {
		return false, err
	}
	return srcHash == dstHash, nil
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

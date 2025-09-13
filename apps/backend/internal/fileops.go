package fileops

import (
	"os"
	"path/filepath"
)

// ListFiles returns a list of files and folders in the given directory.
func ListFiles(dir string) ([]string, error) {
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

// IsDuplicate checks if a file with the same name and size exists in the destination directory.
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
	return srcInfo.Size() == dstInfo.Size(), nil
}

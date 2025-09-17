package fileops

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	IsDirectory bool      `json:"isDirectory"`
	ModTime     time.Time `json:"modTime"`
	CreateTime  time.Time `json:"createTime"`
	Permissions string    `json:"permissions"`
	MimeType    string    `json:"mimeType,omitempty"`
}

type ListOptions struct {
	Depth        int
	Include      []string
	Exclude      []string
	RegexPattern string
	ShowHidden   bool
}

var (
	ErrInvalidPath      = errors.New("invalid path")
	ErrPathNotFound     = errors.New("path not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrPatternInvalid   = errors.New("invalid pattern")
)

func DefaultListOptions() ListOptions {
	return ListOptions{
		Depth:      1,
		ShowHidden: false,
	}
}

func getCreateTime(info fs.FileInfo) time.Time {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		switch runtime.GOOS {
		case "linux":
			return time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
		case "darwin", "windows":
			return time.Unix(0, 0)
		}
	}
	return info.ModTime()
}

func createFileInfo(d fs.DirEntry, path string) (FileInfo, error) {
	info, err := d.Info()
	if err != nil {
		return FileInfo{}, err
	}

	fileInfo := FileInfo{
		Name:        d.Name(),
		Path:        path,
		Size:        info.Size(),
		IsDirectory: d.IsDir(),
		ModTime:     info.ModTime(),
		Permissions: info.Mode().String(),
		CreateTime:  getCreateTime(info),
	}

	if !d.IsDir() {
		if mime, err := DetectMimeType(path); err == nil {
			fileInfo.MimeType = mime
		}
	}

	return fileInfo, nil
}

func shouldIncludeFile(name string, isDir bool, opts ListOptions, depth int) bool {
	// Skip hidden files unless ShowHidden is true
	if !opts.ShowHidden && strings.HasPrefix(name, ".") {
		return false
	}

	// Apply exclude patterns to both files and directories
	for _, pattern := range opts.Exclude {
		if matched, _ := filepath.Match(pattern, name); matched {
			return false
		}
	}

	// For non-directories, apply include patterns and regex
	if !isDir {
		// Apply include patterns
		if len(opts.Include) > 0 {
			included := false
			for _, pattern := range opts.Include {
				if matched, _ := filepath.Match(pattern, name); matched {
					included = true
					break
				}
			}
			if !included {
				return false
			}
		}

		// Apply regex pattern if set
		if opts.RegexPattern != "" {
			re, err := regexp.Compile(opts.RegexPattern)
			if err == nil && !re.MatchString(name) {
				return false
			}
		}
	}

	return true
}

func ListFiles(root string, opts ListOptions) ([]FileInfo, error) {
	if root == "" {
		return nil, ErrInvalidPath
	}

	root = filepath.Clean(root)

	// Validate regex pattern
	if opts.RegexPattern != "" {
		if _, err := regexp.Compile(opts.RegexPattern); err != nil {
			return nil, ErrPatternInvalid
		}
	}

	_, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrPathNotFound
		}
		if os.IsPermission(err) {
			return nil, ErrPermissionDenied
		}
		return nil, err
	}

	var files []FileInfo
	maxDepth := opts.Depth

	var walkDir func(string, int) error
	walkDir = func(currentPath string, currentDepth int) error {
		if maxDepth > 0 && currentDepth > maxDepth {
			return nil
		}

		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			name := entry.Name()
			isDir := entry.IsDir()

			if !shouldIncludeFile(name, isDir, opts, currentDepth) {
				continue
			}

			fullPath := filepath.Join(currentPath, name)

			// Recurse into subdirectories
			if isDir {
				walkDir(fullPath, currentDepth+1)
			}

			// For pattern-based searches, don't add directories to the result
			if isDir && (len(opts.Include) > 0 || opts.RegexPattern != "") {
				continue
			}

			fileInfo, err := createFileInfo(entry, fullPath)
			if err != nil {
				continue
			}

			files = append(files, fileInfo)
		}
		return nil
	}

	err = walkDir(root, 1)
	if err != nil {
		return nil, err
	}

	return files, nil
}

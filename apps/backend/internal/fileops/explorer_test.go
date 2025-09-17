package fileops

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func setup(t *testing.T) string {
	root, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}

	// Create test directory structure
	dirs := []string{
		"dir1",
		"dir1/dir2",
		"dir3",
	}

	for _, dir := range dirs {
		path := filepath.Join(root, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			os.RemoveAll(root)
			t.Fatal(err)
		}
	}

	files := []string{
		"file1.txt",
		"file2.jpg",
		".hidden",
		"dir1/file3.txt",
		"dir1/dir2/file4.txt",
		"dir3/file5.jpg",
	}

	for _, file := range files {
		path := filepath.Join(root, file)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			os.RemoveAll(root)
			t.Fatal(err)
		}
	}

	return root
}

func cleanup(root string) {
	os.RemoveAll(root)
}

func TestListFiles(t *testing.T) {
	root := setup(t)
	defer cleanup(root)

	tests := []struct {
		name          string
		opts          ListOptions
		expectedFiles []string
		check         func([]FileInfo) bool
	}{
		{
			name:          "default options",
			opts:          DefaultListOptions(),
			expectedFiles: []string{"dir1", "dir3", "file1.txt", "file2.jpg"},
			check: func(files []FileInfo) bool {
				return true
			},
		},
		{
			name: "show hidden",
			opts: ListOptions{
				Depth:      1,
				ShowHidden: true,
			},
			expectedFiles: []string{".hidden", "dir1", "dir3", "file1.txt", "file2.jpg"},
			check: func(files []FileInfo) bool {
				return true
			},
		},
		{
			name: "depth 2",
			opts: ListOptions{
				Depth: 2,
			},
			expectedFiles: []string{
				"dir1", "dir1/dir2", "dir1/file3.txt",
				"dir3", "dir3/file5.jpg", "file1.txt", "file2.jpg",
			},
			check: func(files []FileInfo) bool {
				return true
			},
		},
		{
			name: "include pattern",
			opts: ListOptions{
				Depth:   -1,
				Include: []string{"*.jpg"},
			},
			expectedFiles: []string{"dir3/file5.jpg", "file2.jpg"},
			check: func(files []FileInfo) bool {
				return true
			},
		},
		{
			name: "exclude pattern",
			opts: ListOptions{
				Depth:      -1,
				Exclude:    []string{"*.jpg"},
				ShowHidden: true,
			},
			expectedFiles: []string{".hidden", "dir1", "dir1/dir2", "dir1/dir2/file4.txt", "dir1/file3.txt", "dir3", "file1.txt"},
			check: func(files []FileInfo) bool {
				return true
			},
		},
		{
			name: "regex pattern",
			opts: ListOptions{
				Depth:        -1,
				RegexPattern: "file[1-3]",
			},
			expectedFiles: []string{"dir1/file3.txt", "file1.txt", "file2.jpg"},
			check: func(files []FileInfo) bool {
				return true
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files, err := ListFiles(root, tc.opts)
			if err != nil {
				t.Fatal(err)
			}

			var actualFiles []string
			for _, f := range files {
				rel, err := filepath.Rel(root, f.Path)
				if err != nil {
					t.Fatal(err)
				}
				actualFiles = append(actualFiles, rel)
			}
			sort.Strings(actualFiles)
			sort.Strings(tc.expectedFiles)

			if len(actualFiles) != len(tc.expectedFiles) {
				t.Errorf("expected files length %d, got %d\nExpected: %v\nGot: %v",
					len(tc.expectedFiles), len(actualFiles), tc.expectedFiles, actualFiles)
			}

			for i := 0; i < len(actualFiles); i++ {
				if actualFiles[i] != tc.expectedFiles[i] {
					t.Errorf("file at position %d differs: expected %s, got %s",
						i, tc.expectedFiles[i], actualFiles[i])
				}
			}

			if !tc.check(files) {
				t.Error("check failed")
			}
		})
	}
}

func TestListFilesErrors(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		opts          ListOptions
		expectedError error
	}{
		{
			name:          "empty path",
			path:          "",
			opts:          DefaultListOptions(),
			expectedError: ErrInvalidPath,
		},
		{
			name:          "non-existent path",
			path:          "/non/existent/path",
			opts:          DefaultListOptions(),
			expectedError: ErrPathNotFound,
		},
		{
			name: "invalid regex",
			path: "/tmp",
			opts: ListOptions{
				RegexPattern: "[invalid",
			},
			expectedError: ErrPatternInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ListFiles(tc.path, tc.opts)
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

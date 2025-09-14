package fileops

import (
"os"
"path/filepath"
"regexp"
"testing"
)

func setup(t *testing.T) string {
root, err := os.MkdirTemp("", "testdir")
if err != nil {
t.Fatal(err)
}

// Create test directory structure
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
dir := filepath.Dir(path)
if err := os.MkdirAll(dir, 0755); err != nil {
os.RemoveAll(root)
t.Fatal(err)
}
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
expectedCount int
check         func([]FileInfo) bool
}{
{
name:          "default options",
opts:          DefaultListOptions(),
expectedCount: 3, // file1.txt, file2.jpg, dir1
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
expectedCount: 4, // file1.txt, file2.jpg, .hidden, dir1
check: func(files []FileInfo) bool {
hasHidden := false
for _, f := range files {
if filepath.Base(f.Path) == ".hidden" {
hasHidden = true
break
}
}
return hasHidden
},
},
{
name: "depth 2",
opts: ListOptions{
Depth: 2,
},
expectedCount: 5, // file1.txt, file2.jpg, dir1, file3.txt, dir2
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
expectedCount: 2, // file2.jpg, file5.jpg
check: func(files []FileInfo) bool {
for _, f := range files {
if filepath.Ext(f.Path) != ".jpg" {
return false
}
}
return true
},
},
{
name: "exclude pattern",
opts: ListOptions{
Depth:   -1,
Exclude: []string{"*.jpg"},
},
expectedCount: 3, // file1.txt, file3.txt, file4.txt
check: func(files []FileInfo) bool {
for _, f := range files {
if filepath.Ext(f.Path) == ".jpg" {
return false
}
}
return true
},
},
{
name: "regex pattern",
opts: ListOptions{
Depth:        -1,
RegexPattern: "file[1-3]",
},
expectedCount: 3, // file1.txt, file2.jpg, file3.txt
check: func(files []FileInfo) bool {
re := regexp.MustCompile("file[1-3]")
for _, f := range files {
if !re.MatchString(filepath.Base(f.Path)) {
return false
}
}
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

if len(files) != tc.expectedCount {
t.Errorf("expected %d files, got %d", tc.expectedCount, len(files))
}

if !tc.check(files) {
t.Error("check failed")
}

// Basic validation for all files
for _, f := range files {
if f.Name == "" {
t.Error("file name is empty")
}
if f.Path == "" {
t.Error("file path is empty")
}
if !f.IsDirectory && f.Size == 0 {
t.Error("file size is 0")
}
if f.ModTime.IsZero() {
t.Error("mod time is zero")
}
if f.Permissions == "" {
t.Error("permissions is empty")
}
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

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
case "darwin":
return time.Unix(0, 0)
case "windows":
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
Name:        info.Name(),
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

func ListFiles(root string, opts ListOptions) ([]FileInfo, error) {
if root == "" {
return nil, ErrInvalidPath
}

var re *regexp.Regexp
var err error
if opts.RegexPattern != "" {
re, err = regexp.Compile(opts.RegexPattern)
if err != nil {
return nil, ErrPatternInvalid
}
}

var files []FileInfo

entries, err := os.ReadDir(root)
if err != nil {
if os.IsNotExist(err) {
return nil, ErrPathNotFound
}
if os.IsPermission(err) {
return nil, ErrPermissionDenied
}
return nil, err
}

for _, entry := range entries {
name := entry.Name()
path := filepath.Join(root, name)

if !opts.ShowHidden && strings.HasPrefix(name, ".") {
continue
}

excluded := false
for _, pattern := range opts.Exclude {
if matched, _ := filepath.Match(pattern, name); matched {
excluded = true
break
}
}
if excluded {
continue
}

if len(opts.Include) > 0 {
included := false
for _, pattern := range opts.Include {
if matched, _ := filepath.Match(pattern, name); matched {
included = true
break
}
}
if !included && !entry.IsDir() {
continue
}
}

if re != nil && !entry.IsDir() && !re.MatchString(name) {
continue
}

if !entry.IsDir() || opts.Depth != 0 {
fileInfo, err := createFileInfo(entry, path)
if err != nil {
continue
}
files = append(files, fileInfo)
}
}

if opts.Depth == 0 {
return files, nil
}

err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
if err != nil {
if os.IsPermission(err) {
return filepath.SkipDir
}
return err
}

if path == root {
return nil
}

relPath, err := filepath.Rel(root, path)
if err != nil {
return err
}

depth := strings.Count(relPath, string(os.PathSeparator))

if opts.Depth >= 0 && depth > opts.Depth {
if d.IsDir() {
return filepath.SkipDir
}
return nil
}

if !opts.ShowHidden && strings.HasPrefix(filepath.Base(path), ".") {
if d.IsDir() {
return filepath.SkipDir
}
return nil
}

for _, pattern := range opts.Exclude {
if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
if d.IsDir() {
return filepath.SkipDir
}
return nil
}
}

if !d.IsDir() {
if len(opts.Include) > 0 {
matched := false
for _, pattern := range opts.Include {
if m, _ := filepath.Match(pattern, filepath.Base(path)); m {
matched = true
break
}
}
if !matched {
return nil
}
}

if re != nil && !re.MatchString(filepath.Base(path)) {
return nil
}
}

if (d.IsDir() && depth == 1) || (!d.IsDir() && depth <= opts.Depth) {
fileInfo, err := createFileInfo(d, path)
if err != nil {
return nil
}
files = append(files, fileInfo)
}

return nil
})

if err != nil {
if os.IsNotExist(err) {
return nil, ErrPathNotFound
}
if os.IsPermission(err) {
return nil, ErrPermissionDenied
}
return nil, err
}

return files, nil
}

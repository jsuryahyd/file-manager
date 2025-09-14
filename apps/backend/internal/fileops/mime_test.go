package fileops

import (
"os"
"path/filepath"
"strings"
"testing"
)

func TestDetectMimeType(t *testing.T) {
tests := []struct {
name        string
content     string
ext         string
wantMime    string
wantErr     bool
setupFile   bool
}{
{
name:      "text file",
content:   "Hello, World!",
ext:      ".txt",
wantMime: "text/plain",
setupFile: true,
},
{
name:      "html file",
content:   "<html><body>Hello</body></html>",
ext:      ".html",
wantMime: "text/html",
setupFile: true,
},
{
name:      "non-existent file",
ext:      ".txt",
wantMime: "",
wantErr:  true,
setupFile: false,
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
var testPath string
if tc.setupFile {
testFile, err := os.CreateTemp("", "*"+tc.ext)
if err != nil {
t.Fatal(err)
}
testPath = testFile.Name()
defer os.Remove(testPath)

if _, err := testFile.WriteString(tc.content); err != nil {
t.Fatal(err)
}
testFile.Close()
} else {
testPath = filepath.Join(os.TempDir(), "nonexistent"+tc.ext)
}

gotMime, err := DetectMimeType(testPath)
if (err != nil) != tc.wantErr {
t.Errorf("DetectMimeType() error = %v, wantErr %v", err, tc.wantErr)
return
}
if err == nil && !strings.HasPrefix(gotMime, tc.wantMime) {
t.Errorf("DetectMimeType() = %v, want %v", gotMime, tc.wantMime)
}
})
}
}

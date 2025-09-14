package fileops

import (
"net/http"
"os"
)

func DetectMimeType(path string) (string, error) {
file, err := os.Open(path)
if err != nil {
return "", err
}
defer file.Close()

buffer := make([]byte, 512)
n, err := file.Read(buffer)
if err != nil {
return "", err
}

return http.DetectContentType(buffer[:n]), nil
}

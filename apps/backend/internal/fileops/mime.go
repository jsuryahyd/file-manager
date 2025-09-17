package fileops

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

var (
	// Common mime types that might not be registered
	customMimeTypes = map[string]string{
		".md":    "text/markdown",
		".go":    "text/x-go",
		".py":    "text/x-python",
		".js":    "application/javascript",
		".ts":    "application/typescript",
		".jsx":   "text/jsx",
		".tsx":   "text/tsx",
		".vue":   "text/vue",
		".json":  "application/json",
		".yaml":  "application/x-yaml",
		".yml":   "application/x-yaml",
		".toml":  "application/toml",
		".ini":   "text/plain",
		".conf":  "text/plain",
		".sh":    "text/x-shellscript",
		".bash":  "text/x-shellscript",
		".zsh":   "text/x-shellscript",
		".fish":  "text/x-shellscript",
		".sql":   "text/x-sql",
		".c":     "text/x-c",
		".h":     "text/x-c",
		".cpp":   "text/x-c++",
		".hpp":   "text/x-c++",
		".cs":    "text/x-csharp",
		".java":  "text/x-java",
		".scala": "text/x-scala",
		".rs":    "text/x-rust",
		".rb":    "text/x-ruby",
		".php":   "text/x-php",
		".pl":    "text/x-perl",
		".swift": "text/x-swift",
		".kt":    "text/x-kotlin",
		".ex":    "text/x-elixir",
		".exs":   "text/x-elixir",
		".hs":    "text/x-haskell",
		".lua":   "text/x-lua",
		".r":     "text/x-r",
		".jl":    "text/x-julia",
	}
)

// DetectMimeType returns the MIME type of a file based on its extension and content
func DetectMimeType(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext := filepath.Ext(path)

	// First try custom types
	if mimeType, ok := customMimeTypes[ext]; ok {
		return mimeType, nil
	}

	// Try to detect from extension using standard library
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		return mimeType, nil
	}

	// As a last resort, try to detect from file content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer[:n]), nil
}
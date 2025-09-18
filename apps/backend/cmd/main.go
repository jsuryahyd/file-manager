package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"file-manager-backend/internal/config"
	"file-manager-backend/internal/db"
	"file-manager-backend/internal/fileops"
)

// FileEntry defines the structure for file and directory metadata.
type FileEntry struct {
	Name    string      `json:"name"`
	IsDir   bool        `json:"isDir"`
	Size    int64       `json:"size"`
	ModTime string      `json:"modTime"`
}

// SyncRequest defines the structure for a synchronization request.
type SyncRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	dbPath := cfg.Database.Path
	sqlPath := cfg.Database.SQLInit

	dbConn, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err = db.Migrate(dbConn, sqlPath)
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		fmt.Println("Database initialized.")
	} else {
		fmt.Println("Database already exists.")
	}

	filesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		if path == "" {
			path = "." // default to current directory
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var fileEntries []FileEntry
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue // Skip files we can't get info for
			}
			fileEntries = append(fileEntries, FileEntry{
				Name:    entry.Name(),
				IsDir:   entry.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().String(),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fileEntries); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	syncHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Source == "" || req.Destination == "" {
			http.Error(w, "source and destination are required", http.StatusBadRequest)
			return
		}

		_, err := fileops.SyncUniqueFiles(dbConn, req.Source, req.Destination)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	http.Handle("/api/files", corsMiddleware(filesHandler))
	http.Handle("/api/sync", corsMiddleware(syncHandler))

	fmt.Println("File Manager Backend API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

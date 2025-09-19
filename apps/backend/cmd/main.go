package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"file-manager-backend/internal/config"
	"file-manager-backend/internal/db"
	"file-manager-backend/internal/fileops"
)

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
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Error getting user home directory: %v", err)
			http.Error(w, "Cannot get user home directory", http.StatusInternalServerError)
			return
		}

		// Prevent directory traversal
		cleanPath := filepath.Clean(path)
		if strings.HasPrefix(cleanPath, "..") {
			log.Printf("Attempted directory traversal: %s", cleanPath)
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		fullPath := ""
		if filepath.IsAbs(cleanPath) {
			fullPath = cleanPath
		} else {
			fullPath = filepath.Join(homeDir, cleanPath)
		}

		log.Printf("Listing files in: %s", fullPath)
		entries, err := fileops.ListFiles(fullPath)
		if err != nil {
			log.Printf("Error listing files in %s: %v", fullPath, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			log.Printf("Error encoding file list response: %v", err)
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
			log.Printf("Error decoding sync request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Source == "" || req.Destination == "" {
			log.Printf("Missing source or destination in sync request")
			http.Error(w, "source and destination are required", http.StatusBadRequest)
			return
		}

		force, _ := strconv.ParseBool(r.URL.Query().Get("force"))

		pair, err := db.GetSyncPair(dbConn, req.Source, req.Destination)
		if err != nil {
			if err == sql.ErrNoRows {
				// New pair
				if force {
					pairID, err := db.CreateSyncPair(dbConn, req.Source, req.Destination)
					if err != nil {
						log.Printf("Error creating sync pair %s -> %s: %v", req.Source, req.Destination, err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					pair = &db.SyncPair{ID: pairID, SourceDir: req.Source, DestDir: req.Destination}
					log.Printf("New sync pair created: %s -> %s (ID: %d)", pair.SourceDir, pair.DestDir, pair.ID)
				} else {
					log.Printf("New sync pair requires confirmation: %s -> %s", req.Source, req.Destination)
					http.Error(w, "New sync pair requires confirmation", http.StatusConflict)
					return
				}
			} else {
				log.Printf("Error getting sync pair %s -> %s: %v", req.Source, req.Destination, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Existing sync pair found: %s -> %s (ID: %d)", pair.SourceDir, pair.DestDir, pair.ID)
		}

		_, err = fileops.SyncUniqueFiles(dbConn, pair.SourceDir, pair.DestDir, pair.ID)
		if err != nil {
			log.Printf("Error syncing files %s -> %s: %v", pair.SourceDir, pair.DestDir, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Sync completed successfully for pair: %s -> %s", pair.SourceDir, pair.DestDir)
		w.WriteHeader(http.StatusNoContent)
	})

	http.Handle("/api/files/list", corsMiddleware(filesHandler))
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

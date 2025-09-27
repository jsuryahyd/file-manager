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
	"file-manager-backend/internal/sync"
)

// SyncRequest defines the structure for a synchronization request.
type SyncRequest struct {
	Source            string   `json:"source"`
	Destination       string   `json:"destination"`
	CheckDuplicates   bool     `json:"checkDuplicates"`
	OverwriteExisting bool     `json:"overwriteExisting"`
	Recursive         bool     `json:"recursive"`
	SkipPatterns      []string `json:"skipPatterns"`
}

// DeleteRequest defines the structure for a delete request.
type DeleteRequest struct {
	Paths []string `json:"paths"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	dbPath := cfg.Database.Path
	sqlPath := cfg.Database.SQLInit
	migrationsPath := filepath.Join(filepath.Dir(dbPath), "migrations")

	dbConn, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	err = db.Migrate(dbConn, sqlPath, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database migration completed.")

	homeDirHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Error getting user home directory: %v", err)
			http.Error(w, "Cannot get user home directory", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"homeDir": homeDir}); err != nil {
			log.Printf("Error encoding home directory response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	filesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		fileType := r.URL.Query().Get("type")

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
		entries, err := fileops.ListFiles(fullPath, fileType)
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

	syncPreviewHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		opts := sync.Options{
			Recursive:       req.Recursive,
			SkipPatterns:    req.SkipPatterns,
			Overwrite:       req.OverwriteExisting,
			CheckDuplicates: req.CheckDuplicates,
		}

		sourcePath := filepath.FromSlash(req.Source)
		destPath := filepath.FromSlash(req.Destination)

		job, err := sync.NewSyncJob(dbConn, sourcePath, destPath, 0, opts) // syncPairID is 0 because we are not creating a job
		if err != nil {
			log.Printf("Error creating sync job: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result, err := job.Peek()
		if err != nil {
			log.Printf("Error running sync peek: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("Error encoding sync peek response: %v", err)
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

		opts := sync.Options{
			Recursive:       req.Recursive,
			SkipPatterns:    req.SkipPatterns,
			Overwrite:       req.OverwriteExisting,
			CheckDuplicates: req.CheckDuplicates,
		}

		sourcePath := filepath.FromSlash(pair.SourceDir)
		destPath := filepath.FromSlash(pair.DestDir)

		job, err := sync.NewSyncJob(dbConn, sourcePath, destPath, pair.ID, opts)
		if err != nil {
			log.Printf("Error creating sync job: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Run the job in a goroutine to make it non-blocking
		go job.Run()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		if err := json.NewEncoder(w).Encode(map[string]int64{"jobId": pair.ID}); err != nil {
			log.Printf("Error encoding job ID response: %v", err)
		}
	})

	syncJobsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}

		jobs, err := db.GetSyncJobs(dbConn, 10)
		if err != nil {
			log.Printf("Error getting sync jobs: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(jobs); err != nil {
			log.Printf("Error encoding sync jobs response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	findDuplicatesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}

		// It's a good idea to have some security checks on the path, similar to filesHandler.
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Error getting user home directory: %v", err)
			http.Error(w, "Cannot get user home directory", http.StatusInternalServerError)
			return
		}
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


		duplicates, err := fileops.FindDuplicates(fullPath)
		if err != nil {
			log.Printf("Error finding duplicates in %s: %v", fullPath, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(duplicates); err != nil {
			log.Printf("Error encoding duplicates response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	deleteDuplicatesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.Paths) == 0 {
			http.Error(w, "paths are required", http.StatusBadRequest)
			return
		}

		err := fileops.DeleteFiles(req.Paths)
		if err != nil {
			log.Printf("Error deleting files: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	http.Handle("/api/files/list", corsMiddleware(filesHandler))
	http.Handle("/api/sync/preview", corsMiddleware(syncPreviewHandler))
	http.Handle("/api/sync", corsMiddleware(syncHandler))
	http.Handle("/api/sync/jobs", corsMiddleware(syncJobsHandler))
	http.Handle("/api/duplicates/find", corsMiddleware(findDuplicatesHandler))
	http.Handle("/api/duplicates/delete", corsMiddleware(deleteDuplicatesHandler))
	http.Handle("/api/user/home", corsMiddleware(homeDirHandler))

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

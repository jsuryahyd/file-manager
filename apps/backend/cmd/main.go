package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"file-manager-backend/internal/config"
	"file-manager-backend/internal/db"
	"file-manager-backend/internal/fileops"
)

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

	http.HandleFunc("/api/list", func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("dir")
		if dir == "" {
			dir = "." // default to current directory
		}

		opts := fileops.DefaultListOptions()

		// Parse query parameters
		if depth := r.URL.Query().Get("depth"); depth != "" {
			if d, err := strconv.Atoi(depth); err == nil {
				opts.Depth = d
			}
		}
		if pattern := r.URL.Query().Get("pattern"); pattern != "" {
			opts.RegexPattern = pattern
		}
		if include := r.URL.Query().Get("include"); include != "" {
			opts.Include = strings.Split(include, ",")
		}
		if exclude := r.URL.Query().Get("exclude"); exclude != "" {
			opts.Exclude = strings.Split(exclude, ",")
		}
		if showHidden := r.URL.Query().Get("hidden"); showHidden == "true" {
			opts.ShowHidden = true
		}

		files, err := fileops.ListFilesMeta(dir, opts)
		if err != nil {
			switch err {
			case fileops.ErrInvalidPath:
				http.Error(w, err.Error(), http.StatusBadRequest)
			case fileops.ErrPathNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			case fileops.ErrPermissionDenied:
				http.Error(w, err.Error(), http.StatusForbidden)
			case fileops.ErrPatternInvalid:
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(files); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {
		src := r.URL.Query().Get("src")
		dst := r.URL.Query().Get("dst")
		if src == "" || dst == "" {
			http.Error(w, "src and dst required", http.StatusBadRequest)
			return
		}
		copied, err := fileops.SyncUniqueFiles(src, dst)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(copied)
	})

	fmt.Println("File Manager Backend API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

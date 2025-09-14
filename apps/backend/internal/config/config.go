package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadEnv loads environment variables from .env file
func loadEnv(projectRoot string) error {
	envFile := filepath.Join(projectRoot, "apps", "backend", ".env")
	file, err := os.Open(envFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" { // Only set if not already set
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

type Config struct {
	Database struct {
		Path    string `json:"path"`
		SQLInit string `json:"sqlInit"`
	} `json:"database"`
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
}

var cfg *Config

// Load reads configuration from the specified environment
// Falls back to default development config if no environment is specified
func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	// Get project root first as it's needed for .env and paths
	projectRoot := getProjectRoot()

	// Load .env file
	if err := loadEnv(projectRoot); err != nil {
		// Don't return error if .env doesn't exist
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	cfg = &Config{}

	// Set defaults with project root
	cfg.Database.Path = filepath.Join(projectRoot, "apps", "backend", "database", "filemanager.db")
	cfg.Database.SQLInit = filepath.Join(projectRoot, "apps", "backend", "database", "init.sql")
	cfg.Server.Port = "8080"

	// Get config file path from environment, default to development
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	configFile := filepath.Join("config", env+".json")

	// Try to read config file if it exists
	if _, err := os.Stat(configFile); err == nil {
		file, err := os.Open(configFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(cfg); err != nil {
			return nil, err
		}
	}

	// Override with environment variables if set
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if sqlInit := os.Getenv("DB_INIT_SQL"); sqlInit != "" {
		cfg.Database.SQLInit = sqlInit
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}

	// If paths from env/config are relative, make them absolute
	if !filepath.IsAbs(cfg.Database.Path) {
		cfg.Database.Path = filepath.Join(projectRoot, cfg.Database.Path)
	}
	if !filepath.IsAbs(cfg.Database.SQLInit) {
		cfg.Database.SQLInit = filepath.Join(projectRoot, cfg.Database.SQLInit)
	}

	return cfg, nil
}

// getProjectRoot returns the absolute path to the project root
func getProjectRoot() string {
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		fmt.Println(root)
		return root
	}
	// For development, assuming we're in apps/backend/cmd
	pwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Navigate up to project root (2 levels up from cmd/)
	return filepath.Clean(filepath.Join(pwd, "..", ".."))
}

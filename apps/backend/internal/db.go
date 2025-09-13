package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database and returns the connection.
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate runs the database initialization SQL script.
func Migrate(db *sql.DB, sqlPath string) error {
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(content))
	return err
}

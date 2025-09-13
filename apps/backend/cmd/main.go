
package main

import (
	"fmt"
	"log"
	"os"

	"file-manager-backend/internal"
)

func main() {
	dbPath := "../../database/filemanager.db"
	sqlPath := "../../database/init.sql"

	db, err := internal.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err = internal.Migrate(db, sqlPath)
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		fmt.Println("Database initialized.")
	} else {
		fmt.Println("Database already exists.")
	}

	fmt.Println("File Manager Backend: Ready to implement core logic.")
}

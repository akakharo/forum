package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB opens the SQLite database and runs the schema migration.
func InitDB(dbPath string, schemaPath string) {
	var err error

	// Open the SQLite database file (creates it if it doesn't exist)
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Enable foreign key constraints
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Printf("Warning: Failed to enable foreign keys: %v", err)
	}

	// Enable WAL mode for better concurrency and transaction handling
	_, err = DB.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		log.Printf("Warning: Failed to enable WAL mode: %v", err)
	}

	// Read the schema.sql file
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	// Execute the schema SQL to create tables
	_, err = DB.Exec(string(schema))
	if err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	fmt.Println("Database initialized and schema migrated.")
}

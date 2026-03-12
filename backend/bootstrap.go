package backend

import (
	"database/sql"
	_ "embed" // Required for embedding
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/schema.sql
var schema string

var Q *Queries
var DB *sql.DB

func Connect(dbPath string) error {
	var err error

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("Failed to open SQLite: %w", err)
	}

	_, err = DB.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA foreign_keys=ON;
	`)
	if err != nil {
		return fmt.Errorf("Failed to set PRAGMAs: %w", err)
	}

	// Initialize the sqlc-generated queries
	Q = New(DB)

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func ApplySchema() error {
	_, err := DB.Exec(schema)
		if err != nil {
			return fmt.Errorf("Error applying schema: %w", err)
		}

	return nil
}

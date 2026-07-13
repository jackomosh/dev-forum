package sqlite

import (
	"database/sql"
	"fmt"
	"os"

	// Import the go-sqlite3 driver implicitly so database/sql can register it
	_ "github.com/mattn/go-sqlite3"
)

type ClientOptions struct {
	Path            string
	ForeignKeys     bool
	BusyTimeoutMS   int
	MaxOpenConns    int
	MaxIdleConns    int
	SchemaPath      string
	RunMigrations   bool
	MigrationsTable string
	ConnectionName  string
}

type MigrationRecord struct {
	ID        int64
	Name      string
	Checksum  string
	AppliedAt string
}

// InitDB initializes an SQLite connection based on your ClientOptions preferences.
func InitDB(opts ClientOptions) (*sql.DB, error) {
	// 1. Connect to the SQLite file using your path option
	db, err := sql.Open("sqlite3", opts.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}

	// 2. Set limits if specified in your setup
	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}

	// 3. Enforce foreign keys dynamically if explicitly enabled
	if opts.ForeignKeys {
		_, err = db.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to activate foreign key parameters: %w", err)
		}
	}

	// Ping connection to confirm database file handle is active and writable
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database connectivity check failed: %w", err)
	}

	// 4. Read the production schema script from disk
	schemaQuery, err := os.ReadFile(opts.SchemaPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	// 5. Execute table creations
	_, err = db.Exec(string(schemaQuery))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to build tables from schema: %w", err)
	}

	return db, nil
}

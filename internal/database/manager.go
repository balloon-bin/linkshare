// Package database provides all database interactions for linkshare.
// This includes functions to read and write structured link data, setting and
// getting configurations, updating and initializing the schema and backing up
// data
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const expectedSchemaVersion = 1

// DB represents a database connection
type DB struct {
	conn *sql.DB
}

var (
	ErrNotInitialized     = errors.New("database not initialized")
	ErrAlreadyInitialized = errors.New("database already initialized")
	ErrSchemaOutdated     = errors.New("database schema needs updating")
	ErrSchemaUnsupported  = errors.New("database schema is too new for the server")
	ErrMigrationFailed    = errors.New("migration failed")
)

// Open opens a connection to the sqlite database at the given path
func Open(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	conn.SetMaxOpenConns(1) // SQLite only supports one writer at a time

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	_, err = conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign key constraints: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection if it's open
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Initialize the database schema
func (db *DB) Initialize(schemaPath string) error {
	err := db.CheckInitialized()
	if err == nil {
		return ErrAlreadyInitialized
	}

	currentSchema := filepath.Join(schemaPath, "current.sql")
	schema, err := os.ReadFile(currentSchema)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.conn.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	return nil
}

// CheckInitialized returns nil if the database is initialized and an error otherwise
func (db *DB) CheckInitialized() error {
	var count int
	err := db.conn.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='settings'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if database is initialized: %w", err)
	}

	if count == 0 {
		return ErrNotInitialized
	}
	return nil
}

// GetSchemaVersion returns the schema version or an error
func (db *DB) GetSchemaVersion() (int, error) {
	var version string
	err := db.conn.QueryRow("SELECT value FROM settings WHERE key='schema-version'").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	versionInt, err := strconv.Atoi(version)
	if err != nil {
		return 0, fmt.Errorf("invalid schema version: %w", err)
	}

	if versionInt < 1 {
		return 0, fmt.Errorf("invalid schema version %d", versionInt)
	}

	return versionInt, nil
}

// CheckSchemaVersion verifies that the schema is initialized and has the correct version
func (db *DB) CheckSchemaVersion() error {
	err := db.CheckInitialized()
	if err != nil {
		return err
	}
	version, err := db.GetSchemaVersion()
	if err != nil {
		return err
	}
	if version < expectedSchemaVersion {
		return ErrSchemaOutdated
	} else if version > expectedSchemaVersion {
		return ErrSchemaUnsupported
	}
	return nil
}

func (db *DB) transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

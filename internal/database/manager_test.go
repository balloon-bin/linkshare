package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"git.omicron.one/omicron/linkshare/internal/version"
)

func TestOpenClose(t *testing.T) {
	// Create temp file for database
	tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test opening
	db, err := Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Test closing
	err = db.Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	// Create temp directory for test data
	tempDir, err := os.MkdirTemp("", "linkshare-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create schema directory and current.sql file
	schemaDir := filepath.Join(tempDir, "schema")
	err = os.Mkdir(schemaDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create schema directory: %v", err)
	}

	// Write test schema to file
	schemaContent := `CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    kind TEXT CHECK(kind IN ('int', 'string', 'bool', 'json', 'glob')) NOT NULL
);

INSERT INTO settings (key, value, kind) VALUES ('schema-version', '1', 'int');`

	err = os.WriteFile(filepath.Join(schemaDir, "current.sql"), []byte(schemaContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Create temp database file
	dbPath := filepath.Join(tempDir, "test.db")

	// Open database
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test initialization
	err = db.Initialize(schemaDir)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test already initialized error
	err = db.Initialize(schemaDir)
	if err != ErrAlreadyInitialized {
		t.Fatalf("Expected ErrAlreadyInitialized, got: %v", err)
	}
}

func TestCheckInitialized(t *testing.T) {
	// Create temp file for database
	tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Open database
	db, err := Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test not initialized
	err = db.CheckInitialized()
	if err != ErrNotInitialized {
		t.Fatalf("Expected ErrNotInitialized, got: %v", err)
	}

	// Initialize the database manually for testing
	_, err = db.conn.Exec("CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL, kind TEXT NOT NULL)")
	if err != nil {
		t.Fatalf("Failed to create settings table: %v", err)
	}

	// Test initialized
	err = db.CheckInitialized()
	if err != nil {
		t.Fatalf("Expected nil error after initialization, got: %v", err)
	}
}

func TestGetSchemaVersion(t *testing.T) {
	// Create temp file for database
	tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Open database
	db, err := Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize the database manually for testing
	_, err = db.conn.Exec("CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL, kind TEXT NOT NULL)")
	if err != nil {
		t.Fatalf("Failed to create settings table: %v", err)
	}

	_, err = db.conn.Exec("INSERT INTO settings (key, value, kind) VALUES ('schema-version', '1', 'int')")
	if err != nil {
		t.Fatalf("Failed to insert schema version: %v", err)
	}

	// Test schema version
	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Fatalf("Failed to get schema version: %v", err)
	}
	if version != 1 {
		t.Fatalf("Expected schema version 1, got: %d", version)
	}

	// Test invalid schema version
	_, err = db.conn.Exec("UPDATE settings SET value = 'invalid' WHERE key = 'schema-version'")
	if err != nil {
		t.Fatalf("Failed to update schema version: %v", err)
	}

	_, err = db.GetSchemaVersion()
	if err == nil {
		t.Fatal("Expected error for invalid schema version, got nil")
	}
}

func TestCheckSchemaVersion(t *testing.T) {
	// Create temp file for database
	tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Open database
	db, err := Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test not initialized
	err = db.CheckSchemaVersion()
	if err != ErrNotInitialized {
		t.Fatalf("Expected ErrNotInitialized, got: %v", err)
	}

	// Initialize the database manually
	_, err = db.conn.Exec("CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL, kind TEXT NOT NULL)")
	if err != nil {
		t.Fatalf("Failed to create settings table: %v", err)
	}

	// Store current schema version
	originalSchemaVersion := version.SchemaVersion
	defer func() {
		// Restore original schema version after test
		version.SchemaVersion = originalSchemaVersion
	}()

	// Test version match
	_, err = db.conn.Exec("INSERT INTO settings (key, value, kind) VALUES ('schema-version', '1', 'int')")
	if err != nil {
		t.Fatalf("Failed to insert schema version: %v", err)
	}
	version.SchemaVersion = 1
	err = db.CheckSchemaVersion()
	if err != nil {
		t.Fatalf("Expected nil error for matching schema versions, got: %v", err)
	}

	// Test outdated version
	version.SchemaVersion = 2
	err = db.CheckSchemaVersion()
	if err != ErrSchemaOutdated {
		t.Fatalf("Expected ErrSchemaOutdated, got: %v", err)
	}

	// Test unsupported version
	version.SchemaVersion = 1
	_, err = db.conn.Exec("UPDATE settings SET value = '2' WHERE key = 'schema-version'")
	if err != nil {
		t.Fatalf("Failed to update schema version: %v", err)
	}
	err = db.CheckSchemaVersion()
	if err != ErrSchemaUnsupported {
		t.Fatalf("Expected ErrSchemaUnsupported, got: %v", err)
	}
}

func TestTransaction(t *testing.T) {
	// Create temp file for database
	tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Open database
	db, err := Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize the database manually for testing
	_, err = db.conn.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test successful transaction
	err = db.Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO test (value) VALUES (?)", "test-value")
		return err
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Verify data was inserted
	var value string
	err = db.conn.QueryRow("SELECT value FROM test WHERE id = 1").Scan(&value)
	if err != nil {
		t.Fatalf("Failed to query test value: %v", err)
	}
	if value != "test-value" {
		t.Fatalf("Expected 'test-value', got: %s", value)
	}

	// Test failed transaction
	err = db.Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO test (value) VALUES (?)", "should-rollback")
		if err != nil {
			return err
		}
		return sql.ErrTxDone // Force rollback
	})
	if err == nil {
		t.Fatal("Expected error from failed transaction, got nil")
	}

	// Verify data was not inserted (rollback worked)
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM test WHERE value = 'should-rollback'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query test count: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected count 0 after rollback, got: %d", count)
	}

	// Test panic in transaction
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()

		_ = db.Transaction(func(tx *sql.Tx) error {
			panic("test panic")
		})
	}()

	if !panicked {
		t.Fatal("Expected panic to be propagated")
	}
}

package database

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"git.omicron.one/omicron/linkshare/internal/version"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()

	db, err := Open(":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, db)

	err = db.Initialize("../../schema")
	assert.NoError(t, err)

	return db
}

func TestOpen(t *testing.T) {
	t.Run("valid database file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		assert.NoError(t, db.conn.Ping())
	})

	t.Run("invalid database path", func(t *testing.T) {
		_, err := Open("/nonexistent/directory/test.db")
		assert.Error(t, err)
	})

	t.Run("foreign keys enabled", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		var fkEnabled int
		err = db.conn.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
		assert.NoError(t, err)
		assert.Equal(t, 1, fkEnabled)
	})
}

func TestClose(t *testing.T) {
	tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	db, err := Open(tempFile.Name())
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	// Double close
	err = db.Close()
	assert.NoError(t, err)
}

func TestInitialize(t *testing.T) {
	t.Run("standard", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		err = db.Initialize("../../schema")
		assert.NoError(t, err)

		// Verify it actually worked by checking for settings table
		err = db.CheckInitialized()
		assert.NoError(t, err)

		// Verify schema version matches expected
		schemaVersion, err := db.GetSchemaVersion()
		assert.NoError(t, err)
		assert.Equal(t, version.SchemaVersion, schemaVersion)
	})

	t.Run("double", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		// Initialize once
		err = db.Initialize("../../schema")
		assert.NoError(t, err)

		// Try to initialize again
		err = db.Initialize("../../schema")
		assert.Equal(t, ErrAlreadyInitialized, err)
	})

	t.Run("missing file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		err = db.Initialize("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read schema file")
	})

	t.Run("corrupted file", func(t *testing.T) {
		// Create temp schema directory with invalid SQL
		tempDir, err := os.MkdirTemp("", "linkshare-schema-test-*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		corruptSchema := "CREATE TABLE invalid syntax here"
		err = os.WriteFile(filepath.Join(tempDir, "current.sql"), []byte(corruptSchema), 0o644)
		assert.NoError(t, err)

		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		err = db.Initialize(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize database")
	})
}

func TestTransaction(t *testing.T) {
	t.Run("successful transaction", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		// Execute transaction that should succeed
		err = db.Transaction(func(tx *sql.Tx) error {
			_, err := tx.Exec("CREATE TABLE test_table (id INTEGER)")
			return err
		})
		assert.NoError(t, err)

		// Verify table was created
		var count int
		err = db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("with error", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		// Execute transaction that should fail
		err = db.Transaction(func(tx *sql.Tx) error {
			_, err := tx.Exec("CREATE TABLE test_table (id INTEGER)")
			if err != nil {
				return err
			}
			return errors.New("forced error")
		})
		assert.Error(t, err)

		// Verify table was not created (rolled back)
		var count int
		err = db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("with panic", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
					assert.Equal(t, "test panic", r)
				}
			}()

			_ = db.Transaction(func(tx *sql.Tx) error {
				_, err := tx.Exec("CREATE TABLE test_table (id INTEGER)")
				if err != nil {
					return err
				}
				panic("test panic")
			})
		}()

		assert.True(t, panicked)

		// Verify table was not created (rolled back)
		var count int
		err = db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestCheckSchemaVersion(t *testing.T) {
	t.Run("not initialized", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "linkshare-test-*.db")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		db, err := Open(tempFile.Name())
		assert.NoError(t, err)
		defer db.Close()

		err = db.CheckSchemaVersion()
		assert.Equal(t, ErrNotInitialized, err)
	})

	t.Run("version match", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		err := db.CheckSchemaVersion()
		assert.NoError(t, err)
	})

	t.Run("outdated schema", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		version.SchemaVersion += 1

		err := db.CheckSchemaVersion()
		assert.Equal(t, ErrSchemaOutdated, err)
	})

	t.Run("unsupported schema", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		// Update schema version to simulate future database
		_, err := db.conn.Exec("UPDATE settings SET value = '99' WHERE key = 'schema-version'")
		assert.NoError(t, err)

		err = db.CheckSchemaVersion()
		assert.Equal(t, ErrSchemaUnsupported, err)
	})

	t.Run("invalid schema version", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		// Update schema version to invalid value
		_, err := db.conn.Exec("UPDATE settings SET value = 'not-a-number' WHERE key = 'schema-version'")
		assert.NoError(t, err)

		err = db.CheckSchemaVersion()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid schema version")
	})

	t.Run("missing schema version key", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		// Delete the schema-version key
		_, err := db.conn.Exec("DELETE FROM settings WHERE key = 'schema-version'")
		assert.NoError(t, err)

		err = db.CheckSchemaVersion()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get schema version")
	})
}

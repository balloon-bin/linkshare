package links_test

import (
	"os"
	"testing"
	"time"

	"git.omicron.one/omicron/linkshare/internal/database"
	"git.omicron.one/omicron/linkshare/internal/database/links"
)

func setupTestDB(t *testing.T) (*database.DB, string) {
	t.Helper()

	cwd, err := os.Getwd()
	t.Logf("Current working directory: %s", cwd)

	// Create temp file for database
	tempFile, err := os.CreateTemp("", "linkshare-links-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()

	dbPath := tempFile.Name()

	// Open database
	db, err := database.Open(dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to open database: %v", err)
	}

	// Initialize database with schema
	err = db.Initialize("../../../schema")
	if err != nil {
		db.Close()
		os.Remove(dbPath)
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return db, dbPath
}

func TestRepository_Create(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	repo := links.NewRepository(db)

	// Test creating a link
	id, err := repo.Create("https://example.com", "Example", false)
	if err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Expected positive ID, got %d", id)
	}

	// Verify link was created by retrieving it
	link, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Failed to get link: %v", err)
	}

	if link.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", link.URL)
	}
}

func TestRepository_Get(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	repo := links.NewRepository(db)

	// Insert test data
	id, err := repo.Create("https://example.com", "Example", true)
	if err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	// Test getting a link
	link, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Failed to get link: %v", err)
	}

	if link.ID != id {
		t.Errorf("Expected ID %d, got %d", id, link.ID)
	}
	if link.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", link.URL)
	}
	if link.Title != "Example" {
		t.Errorf("Expected Title 'Example', got '%s'", link.Title)
	}
	if link.IsPrivate != true {
		t.Errorf("Expected IsPrivate true, got %v", link.IsPrivate)
	}
	if link.UpdatedAt.IsSome() {
		t.Errorf("Expected UpdatedAt to be None, got %v", link.UpdatedAt)
	}

	// Test getting non-existent link
	_, err = repo.Get(id + 1)
	if err == nil {
		t.Fatal("Expected error when getting non-existent link")
	}
}

func TestRepository_Update(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	repo := links.NewRepository(db)

	// Insert test data
	id, err := repo.Create("https://example.com", "Example", false)
	if err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	// Test updating a link
	err = repo.Update(id, "https://updated.com", "Updated", true)
	if err != nil {
		t.Fatalf("Failed to update link: %v", err)
	}

	// Verify link was updated
	link, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Failed to get link: %v", err)
	}

	if link.URL != "https://updated.com" {
		t.Errorf("Expected URL 'https://updated.com', got '%s'", link.URL)
	}
	if link.Title != "Updated" {
		t.Errorf("Expected Title 'Updated', got '%s'", link.Title)
	}
	if link.IsPrivate != true {
		t.Errorf("Expected IsPrivate true, got %v", link.IsPrivate)
	}
	if !link.UpdatedAt.IsSome() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestRepository_Delete(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	repo := links.NewRepository(db)

	// Insert test data
	id, err := repo.Create("https://example.com", "Example", false)
	if err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	// Test deleting a link
	err = repo.Delete(id)
	if err != nil {
		t.Fatalf("Failed to delete link: %v", err)
	}

	// Verify link was deleted
	_, err = repo.Get(id)
	if err == nil {
		t.Fatal("Expected error after deletion")
	}
}

func TestRepository_List(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	repo := links.NewRepository(db)

	// Insert test data
	urls := []struct {
		url       string
		isPrivate bool
	}{
		{"https://example1.com", true},
		{"https://example2.com", false},
		{"https://example3.com", false},
		{"https://example4.com", true},
		{"https://example5.com", false},
	}

	for i, info := range urls {
		_, err := repo.Create(info.url, "Example "+string(rune('A'+i)), info.isPrivate)
		if err != nil {
			t.Fatalf("Failed to create link: %v", err)
		}

		// Add a small delay to ensure different created_at times
		time.Sleep(10 * time.Millisecond)
	}

	// Test full listing with pagination
	links, err := repo.List(true, 0, 3)
	if err != nil {
		t.Fatalf("Failed to list links: %v", err)
	}

	if len(links) != 3 {
		t.Fatalf("Expected 3 links, got %d", len(links))
	}

	// Check order (newest first)
	for i := 0; i < len(links)-1; i++ {
		if links[i].CreatedAt.Before(links[i+1].CreatedAt) {
			t.Errorf("Links not in correct order")
		}
	}

	// Test second page of full listing
	links, err = repo.List(true, 3, 2)
	if err != nil {
		t.Fatalf("Failed to list links: %v", err)
	}

	if len(links) != 2 {
		t.Fatalf("Expected 2 links, got %d", len(links))
	}

	// Test public listing
	links, err = repo.List(false, 0, 3)
	if err != nil {
		t.Fatalf("Failed to list links: %v", err)
	}

	if len(links) != 3 {
		t.Fatalf("Expected 3 links, got %d", len(links))
	}

	for _, link := range links {
		if link.IsPrivate {
			t.Fatalf("private link in public listing %v", link)
		}
	}

	// Try to get more public links
	links, err = repo.List(false, 3, 3)
	if err != nil {
		t.Fatalf("Failed to list links: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("Expected 0 links, got %d", len(links))
	}
}

func TestRepository_Count(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	repo := links.NewRepository(db)

	// Check full count with empty table
	count, err := repo.Count(true)
	if err != nil {
		t.Fatalf("Failed to count links: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 links, got %d", count)
	}

	// Check public count with empty table
	count, err = repo.Count(false)
	if err != nil {
		t.Fatalf("Failed to count links: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 links, got %d", count)
	}

	// Insert test data
	numLinks := 5
	for i := 0; i < numLinks; i++ {
		_, err := repo.Create(
			"https://example"+string(rune('1'+i))+".com",
			"Example "+string(rune('A'+i)),
			i%2 == 1,
		)
		if err != nil {
			t.Fatalf("Failed to create link: %v", err)
		}
	}
	pubLinks := numLinks / 2

	// Check full count again
	count, err = repo.Count(true)
	if err != nil {
		t.Fatalf("Failed to count links: %v", err)
	}
	if count != numLinks {
		t.Fatalf("Expected %d links, got %d", numLinks, count)
	}

	// Check public count again
	count, err = repo.Count(false)
	if err != nil {
		t.Fatalf("Failed to count links: %v", err)
	}
	if count != pubLinks {
		t.Fatalf("Expected %d links, got %d", pubLinks, count)
	}
}

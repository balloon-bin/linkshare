package links

import (
	"database/sql"
	"time"

	"git.omicron.one/omicron/linkshare/internal/database"
	. "git.omicron.one/omicron/linkshare/internal/util/option"
)

// Link represents a stored link
type Link struct {
	ID        int64
	URL       string
	Title     string
	CreatedAt time.Time
	UpdatedAt Option[time.Time]
	IsPrivate bool
}

// Repository handles link storage operations
type Repository struct {
	db *database.DB
}

// NewRepository creates a new link repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Create adds a new link to the database
func (r *Repository) Create(url, title string, isPrivate bool) (int64, error) {
	var id int64
	err := r.db.Transaction(func(tx *sql.Tx) error {
		now := time.Now().UTC().Format(time.RFC3339)
		result, err := tx.Exec(
			"INSERT INTO links (url, title, created_at, is_private) VALUES (?, ?, ?, ?)",
			url, title, now, isPrivate,
		)
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		return err
	})
	return id, err
}

// Get retrieves a single link by ID
func (r *Repository) Get(id int64) (*Link, error) {
	var (
		link      Link
		createdAt string
		updatedAt sql.NullString
	)

	err := r.db.Transaction(func(tx *sql.Tx) error {
		row := tx.QueryRow(
			"SELECT id, url, title, created_at, updated_at, is_private FROM links WHERE id = ?",
			id,
		)

		err := row.Scan(&link.ID, &link.URL, &link.Title, &createdAt, &updatedAt, &link.IsPrivate)
		if err != nil {
			return err
		}

		created, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return err
		}
		link.CreatedAt = created

		if updatedAt.Valid {
			updated, err := time.Parse(time.RFC3339, updatedAt.String)
			if err != nil {
				return err
			}
			link.UpdatedAt = Some(updated)
		} else {
			link.UpdatedAt = None[time.Time]()
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// Update updates an existing link's fields
func (r *Repository) Update(id int64, url, title string, isPrivate bool) error {
	return r.db.Transaction(func(tx *sql.Tx) error {
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := tx.Exec(
			"UPDATE links SET url = ?, title = ?, updated_at = ?, is_private = ? WHERE id = ?",
			url, title, now, isPrivate, id,
		)
		return err
	})
}

// Delete removes a link from the database
func (r *Repository) Delete(id int64) error {
	return r.db.Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM links WHERE id = ?", id)
		return err
	})
}

// List returns a paginated list of links
func (r *Repository) List(includePrivate bool, offset, limit int) ([]*Link, error) {
	var links []*Link

	err := r.db.Transaction(func(tx *sql.Tx) error {
		var rows *sql.Rows
		var err error

		if includePrivate {
			rows, err = tx.Query(
				`SELECT id, url, title, created_at, updated_at, is_private
				FROM links ORDER BY created_at DESC LIMIT ? OFFSET ?`,
				limit, offset,
			)
		} else {
			rows, err = tx.Query(
				`SELECT id, url, title, created_at, updated_at, is_private
				FROM links WHERE is_private = 0 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
				limit, offset,
			)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				link      Link
				createdAt string
				updatedAt sql.NullString
			)

			err := rows.Scan(&link.ID, &link.URL, &link.Title, &createdAt, &updatedAt, &link.IsPrivate)
			if err != nil {
				return err
			}

			created, err := time.Parse(time.RFC3339, createdAt)
			if err != nil {
				return err
			}
			link.CreatedAt = created

			if updatedAt.Valid {
				updated, err := time.Parse(time.RFC3339, updatedAt.String)
				if err != nil {
					return err
				}
				link.UpdatedAt = Some(updated)
			} else {
				link.UpdatedAt = None[time.Time]()
			}

			links = append(links, &link)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return links, nil
}

// Count returns the total number of links in the database
func (r *Repository) Count(includePrivate bool) (int, error) {
	var count int

	err := r.db.Transaction(func(tx *sql.Tx) error {
		var row *sql.Row
		if includePrivate {
			row = tx.QueryRow("SELECT COUNT(*) FROM links")
		} else {
			row = tx.QueryRow("SELECT COUNT(*) FROM links WHERE is_private")
		}
		return row.Scan(&count)
	})

	return count, err
}

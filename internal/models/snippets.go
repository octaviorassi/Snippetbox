package models

import (
	"database/sql"
	"errors"
	"time"
)

type Snippet struct {
	ID		int
	Title 	string
	Content	string
	Created	time.Time
	Expires	time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {

	stmt := `INSERT INTO snippets (title, content, created, expires)
			 VALUES (?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`
	
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil { return 0, err }

	id, err := result.LastInsertId()
	if err != nil { return 0, err }

	return int(id), nil
}

/* Get returns the Snippet identified by `id` if it exists, or an error if it does not */
func (m *SnippetModel) Get(id int) (Snippet, error) {
	
	var s Snippet
	stmt :=	`SELECT id, title, content, created, expires FROM snippets
			 WHERE expires > UTC_TIMESTAMP() AND id = ?`

	err := m.DB.QueryRow(stmt, id).
				Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

	if err != nil {
		// Check if the error is due to not finding any rows matching the ID
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
			 WHERE	expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`


	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// The closing must be deferred after the error-checking because if it was 
	// executed before checking. Otherwise, if an error happens at Query(), 
	// a panic will attempt to close a nil resultset
	// Closing the rows resultset IS NOT OPTIONAL. Failing to do so will keep the
	// underlying db connection open and could exhaust the connection pool rapidly.
	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		var s Snippet
				
		err := rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

		// If any of the scans fails, the whole thing is aborted
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}
	
	// Now we can retrieve any error encountered during the iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}


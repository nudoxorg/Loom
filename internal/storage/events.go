package storage

import (
	"database/sql"
	"fmt"
)

func InsertEvent(db *sql.DB, e Event) error {
	_, err := db.Exec(
		`INSERT INTO events (agent, kind, message, path, timestamp) VALUES (?, ?, ?, ?, ?)`,
		e.Agent, e.Kind, e.Message, e.Path, e.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("inserting event: %w", err)
	}

	return nil
}

func ListEvents(db *sql.DB, limit int) ([]Event, error) {
	rows, err := db.Query(
		`SELECT id, agent, kind, message, path, timestamp FROM events ORDER BY timestamp DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Agent, &e.Kind, &e.Message, &e.Path, &e.Timestamp); err != nil {
			return nil, fmt.Errorf("scanning events: %w", err)
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating events: %w", err)
	}

	return events, nil
}

package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening connection to db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
            agent     TEXT NOT NULL,
            kind      TEXT NOT NULL,
            message   TEXT NOT NULL,
            path      TEXT,
            timestamp DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("migrating db: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS claims (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
          	path        TEXT NOT NULL,
          	agent       TEXT NOT NULL,
          	claimed_at  DATETIME NOT NULL,
          	released_at DATETIME
		)
	`)
	if err != nil {
		return fmt.Errorf("migrating db: %w", err)
	}

	return nil
}

func Close(db *sql.DB) error {
	return db.Close()
}

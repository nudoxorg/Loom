package storage

import (
	"database/sql"
	"fmt"
)

// InsertClaim records path as claimed by c.Agent. If that agent already
// holds an active claim on the path, it refreshes claimed_at instead of
// inserting a duplicate row. It returns whether a new claim was created,
// so callers can decide whether this is event-worthy: a fresh claim is
// (an agent starting on a path is news); a refresh isn't (the agent is
// just re-affirming a claim it already holds).
//
// It deliberately does not dedupe across different agents on the same
// path — a second agent claiming an already-claimed path is exactly the
// collision this table exists to surface, so both rows must stay active.
func InsertClaim(db *sql.DB, c Claim) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("beginning claim transaction: %w", err)
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRow(
		`SELECT id FROM claims WHERE path = ? AND agent = ? AND released_at IS NULL`,
		c.Path, c.Agent,
	).Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		if _, err := tx.Exec(
			`INSERT INTO claims (path, agent, claimed_at) VALUES (?, ?, ?)`,
			c.Path, c.Agent, c.ClaimedAt,
		); err != nil {
			return false, fmt.Errorf("inserting claim: %w", err)
		}
	case err != nil:
		return false, fmt.Errorf("checking existing claim: %w", err)
	default:
		if _, err := tx.Exec(
			`UPDATE claims SET claimed_at = ? WHERE id = ?`,
			c.ClaimedAt, id,
		); err != nil {
			return false, fmt.Errorf("refreshing claim: %w", err)
		}
		return false, tx.Commit()
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("committing claim: %w", err)
	}

	return true, nil
}

func ReleaseClaim(db *sql.DB, path string, agent string) error {
	_, err := db.Exec(
		`UPDATE claims SET released_at = datetime('now') WHERE path = ? AND agent = ? AND released_at IS NULL`,
		path, agent,
	)
	if err != nil {
		return fmt.Errorf("releasing claim: %w", err)
	}

	return nil
}

func ListActiveClaims(db *sql.DB) ([]Claim, error) {
	rows, err := db.Query(
		`SELECT id, path, agent, claimed_at FROM claims WHERE released_at IS NULL`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying claims: %w", err)
	}
	defer rows.Close()

	var claims []Claim
	for rows.Next() {
		var c Claim
		if err := rows.Scan(&c.ID, &c.Path, &c.Agent, &c.ClaimedAt); err != nil {
			return nil, fmt.Errorf("scanning claim: %w", err)
		}
		claims = append(claims, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating claims: %w", err)
	}

	return claims, nil
}

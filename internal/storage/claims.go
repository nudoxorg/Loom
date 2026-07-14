package storage

import (
	"database/sql"
	"fmt"
)

func InsertClaim(db *sql.DB, c Claim) error {
	_, err := db.Exec(
		`INSERT INTO claims (path, agent, claimed_at) VALUES (?, ?, ?)`,
		c.Path, c.Agent, c.ClaimedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting claim: %w", err)
	}

	return nil
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

package storage

import (
	"time"
)

type Event struct {
	ID        int64
	Agent     string
	Kind      string
	Message   string
	Path      string
	Timestamp time.Time
}

type Claim struct {
	ID         int64
	Path       string
	Agent      string
	ClaimedAt  time.Time
	ReleasedAt *time.Time
}

package project

import (
	"path/filepath"
	"sort"

	"github.com/nudoxorg/loom/internal/storage"
)

// LabeledEvent and LabeledClaim pair storage rows with which project
// they came from, for views that span more than one project's db.
type LabeledEvent struct {
	Project string
	Event   storage.Event
}

type LabeledClaim struct {
	Project string
	Claim   storage.Claim
}

// AggregateAll walks every known project plus the global log, returning
// currently active claims and recent events across all of them, merged
// and sorted most-recent-first. Events are truncated to limit; claims
// are not (there are typically few active at once).
func AggregateAll(limit int) ([]LabeledClaim, []LabeledEvent, error) {
	projects, err := ListAll()
	if err != nil {
		return nil, nil, err
	}

	var claims []LabeledClaim
	var events []LabeledEvent

	for _, p := range projects {
		label := filepath.Base(p.Root)

		db, err := storage.Open(p.DBPath)
		if err != nil {
			return nil, nil, err
		}

		projectClaims, err := storage.ListActiveClaims(db)
		if err != nil {
			storage.Close(db)
			return nil, nil, err
		}
		for _, c := range projectClaims {
			claims = append(claims, LabeledClaim{Project: label, Claim: c})
		}

		projectEvents, err := storage.ListEvents(db, limit)
		if err != nil {
			storage.Close(db)
			return nil, nil, err
		}
		for _, e := range projectEvents {
			events = append(events, LabeledEvent{Project: label, Event: e})
		}

		storage.Close(db)
	}

	globalDBPath, err := GlobalDBPath()
	if err != nil {
		return nil, nil, err
	}

	globalDB, err := storage.Open(globalDBPath)
	if err != nil {
		return nil, nil, err
	}

	globalEvents, err := storage.ListEvents(globalDB, limit)
	if err != nil {
		storage.Close(globalDB)
		return nil, nil, err
	}
	for _, e := range globalEvents {
		events = append(events, LabeledEvent{Project: "global", Event: e})
	}
	storage.Close(globalDB)

	sort.Slice(claims, func(i, j int) bool {
		return claims[i].Claim.ClaimedAt.After(claims[j].Claim.ClaimedAt)
	})
	sort.Slice(events, func(i, j int) bool {
		return events[i].Event.Timestamp.After(events[j].Event.Timestamp)
	})
	if len(events) > limit {
		events = events[:limit]
	}

	return claims, events, nil
}

package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nudoxorg/loom/internal/storage"
)

func withAggregateTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if err := EnsureHome(); err != nil {
		t.Fatalf("EnsureHome() error = %v", err)
	}
}

// fakeProject creates a project directory under ~/.loom/projects/<slug>
// with its own loom.db, optionally writing the "root" sidecar file, and
// inserts one event and one active claim into it.
func fakeProject(t *testing.T, slug string, withRoot bool, root string, eventTime time.Time) {
	t.Helper()

	homeDir, err := HomeDir()
	if err != nil {
		t.Fatalf("HomeDir() error = %v", err)
	}

	dir := filepath.Join(homeDir, "projects", slug)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", dir, err)
	}

	if withRoot {
		if err := os.WriteFile(filepath.Join(dir, "root"), []byte(root), 0o600); err != nil {
			t.Fatalf("writing root file: %v", err)
		}
	}

	db, err := storage.Open(filepath.Join(dir, "loom.db"))
	if err != nil {
		t.Fatalf("storage.Open() error = %v", err)
	}
	defer storage.Close(db)

	if err := storage.InsertEvent(db, storage.Event{
		Agent:     "test-agent",
		Kind:      "log",
		Message:   "event in " + slug,
		Timestamp: eventTime,
	}); err != nil {
		t.Fatalf("InsertEvent() error = %v", err)
	}

	if _, err := storage.InsertClaim(db, storage.Claim{
		Path:      "some/path",
		Agent:     "test-agent",
		ClaimedAt: eventTime,
	}); err != nil {
		t.Fatalf("InsertClaim() error = %v", err)
	}
}

func TestAggregateAllMergesAcrossProjectsAndGlobal(t *testing.T) {
	withAggregateTempHome(t)

	now := time.Now()
	fakeProject(t, "Users-alice-code-api-service", true, "/Users/alice/code/api-service", now.Add(-1*time.Hour))
	fakeProject(t, "Users-alice-code-website", true, "/Users/alice/code/website", now.Add(-2*time.Hour))

	globalDBPath, err := GlobalDBPath()
	if err != nil {
		t.Fatalf("GlobalDBPath() error = %v", err)
	}
	globalDB, err := storage.Open(globalDBPath)
	if err != nil {
		t.Fatalf("storage.Open(global) error = %v", err)
	}
	if err := storage.InsertEvent(globalDB, storage.Event{
		Agent:     "manual",
		Kind:      "log",
		Message:   "cross-project note",
		Timestamp: now,
	}); err != nil {
		t.Fatalf("InsertEvent(global) error = %v", err)
	}
	storage.Close(globalDB)

	claims, events, err := AggregateAll(20)
	if err != nil {
		t.Fatalf("AggregateAll() error = %v", err)
	}

	if len(claims) != 2 {
		t.Fatalf("len(claims) = %d, want 2", len(claims))
	}
	if len(events) != 3 {
		t.Fatalf("len(events) = %d, want 3", len(events))
	}

	// Most recent first: global note, then api-service, then website.
	wantOrder := []string{"global", "api-service", "website"}
	for i, want := range wantOrder {
		if events[i].Project != want {
			t.Fatalf("events[%d].Project = %q, want %q", i, events[i].Project, want)
		}
	}
}

func TestAggregateAllTruncatesEventsToLimit(t *testing.T) {
	withAggregateTempHome(t)

	now := time.Now()
	fakeProject(t, "Users-alice-code-api-service", true, "/Users/alice/code/api-service", now.Add(-1*time.Hour))
	fakeProject(t, "Users-alice-code-website", true, "/Users/alice/code/website", now.Add(-2*time.Hour))

	_, events, err := AggregateAll(1)
	if err != nil {
		t.Fatalf("AggregateAll() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Project != "api-service" {
		t.Fatalf("events[0].Project = %q, want %q (most recent)", events[0].Project, "api-service")
	}
}

func TestAggregateAllFallsBackToSlugWithoutRootFile(t *testing.T) {
	withAggregateTempHome(t)

	fakeProject(t, "Users-alice-code-legacy-project", false, "", time.Now())

	_, events, err := AggregateAll(20)
	if err != nil {
		t.Fatalf("AggregateAll() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Project != "Users-alice-code-legacy-project" {
		t.Fatalf("events[0].Project = %q, want raw slug fallback", events[0].Project)
	}
}

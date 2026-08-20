package config

import (
	"testing"

	"github.com/nudoxorg/loom/internal/project"
)

// withTempHome points HomeDir() at a scratch directory so tests never touch
// the real ~/.loom/settings.json, mirroring the project.EnsureHome() call
// every CLI command makes before touching config.
func withTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if err := project.EnsureHome(); err != nil {
		t.Fatalf("EnsureHome() error = %v", err)
	}
}

func TestLoadReturnsDefaultsWhenFileMissing(t *testing.T) {
	withTempHome(t)

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if settings != Default() {
		t.Fatalf("Load() = %+v, want %+v", settings, Default())
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withTempHome(t)

	want := Settings{DefaultAgent: "claude-code", DefaultLimit: 50}
	if err := Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
}

func TestGetKnownKeys(t *testing.T) {
	s := Settings{DefaultAgent: "cursor", DefaultLimit: 42}

	if v, err := Get(s, "default_agent"); err != nil || v != "cursor" {
		t.Fatalf("Get(default_agent) = %q, %v; want %q, nil", v, err, "cursor")
	}

	if v, err := Get(s, "default_limit"); err != nil || v != "42" {
		t.Fatalf("Get(default_limit) = %q, %v; want %q, nil", v, err, "42")
	}
}

func TestGetUnknownKey(t *testing.T) {
	if _, err := Get(Default(), "nope"); err == nil {
		t.Fatal("Get(nope) error = nil, want error")
	}
}

func TestSetKnownKeys(t *testing.T) {
	s := Default()

	if err := Set(&s, "default_agent", "claude-code"); err != nil {
		t.Fatalf("Set(default_agent) error = %v", err)
	}
	if s.DefaultAgent != "claude-code" {
		t.Fatalf("DefaultAgent = %q, want %q", s.DefaultAgent, "claude-code")
	}

	if err := Set(&s, "default_limit", "100"); err != nil {
		t.Fatalf("Set(default_limit) error = %v", err)
	}
	if s.DefaultLimit != 100 {
		t.Fatalf("DefaultLimit = %d, want %d", s.DefaultLimit, 100)
	}
}

func TestSetInvalidLimit(t *testing.T) {
	s := Default()
	if err := Set(&s, "default_limit", "not-a-number"); err == nil {
		t.Fatal("Set(default_limit, not-a-number) error = nil, want error")
	}
}

func TestSetUnknownKey(t *testing.T) {
	s := Default()
	if err := Set(&s, "nope", "value"); err == nil {
		t.Fatal("Set(nope) error = nil, want error")
	}
}

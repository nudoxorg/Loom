package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallWritesScriptsAndSettings(t *testing.T) {
	root := t.TempDir()

	changed, err := Install(root)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !changed {
		t.Fatalf("Install() changed = false, want true on first run")
	}

	for _, name := range []string{sessionStartScript, preToolUseScript} {
		path := filepath.Join(root, ".claude", "hooks", name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable: mode = %v", path, info.Mode())
		}
	}

	settings := readSettings(t, root)
	if len(settings["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d, want 1", len(settings["PreToolUse"]))
	}
	if settings["PreToolUse"][0].Matcher != "Write|Edit|MultiEdit" {
		t.Fatalf("PreToolUse matcher = %q, want Write|Edit|MultiEdit", settings["PreToolUse"][0].Matcher)
	}
	if len(settings["SessionStart"]) != 1 {
		t.Fatalf("SessionStart groups = %d, want 1", len(settings["SessionStart"]))
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	root := t.TempDir()

	if _, err := Install(root); err != nil {
		t.Fatalf("first Install() error = %v", err)
	}

	changed, err := Install(root)
	if err != nil {
		t.Fatalf("second Install() error = %v", err)
	}
	if changed {
		t.Fatalf("second Install() changed = true, want false")
	}

	settings := readSettings(t, root)
	if len(settings["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d after second install, want 1 (no duplicate)", len(settings["PreToolUse"]))
	}
}

func TestInstallPreservesExistingSettings(t *testing.T) {
	root := t.TempDir()

	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	existing := `{
		"model": "opus",
		"hooks": {
			"Stop": [
				{"hooks": [{"type": "command", "command": "some-other-tool"}]}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("writing existing settings.json: %v", err)
	}

	if _, err := Install(root); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing settings.json: %v", err)
	}

	var model string
	if err := json.Unmarshal(raw["model"], &model); err != nil {
		t.Fatalf("parsing model: %v", err)
	}
	if model != "opus" {
		t.Fatalf("model = %q, want %q (must survive install)", model, "opus")
	}

	settings := readSettings(t, root)
	if len(settings["Stop"]) != 1 || settings["Stop"][0].Hooks[0].Command != "some-other-tool" {
		t.Fatalf("existing Stop hook was not preserved: %+v", settings["Stop"])
	}
	if len(settings["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d, want 1", len(settings["PreToolUse"]))
	}
}

func readSettings(t *testing.T, root string) map[string][]hookMatcher {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing settings.json: %v", err)
	}

	var hooksByEvent map[string][]hookMatcher
	if err := json.Unmarshal(raw["hooks"], &hooksByEvent); err != nil {
		t.Fatalf("parsing hooks: %v", err)
	}

	return hooksByEvent
}

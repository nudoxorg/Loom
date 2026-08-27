// Antigravity's PreToolUse decision/reason surfacing behavior (see the
// UNVERIFIED comment on antigravityScriptContent) depends on Antigravity's
// live runtime, not just its documented schema — that can't be captured
// as a unit test here. These tests only cover what Loom itself controls:
// the files it writes and the merge/idempotency guarantees around them.
package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAntigravityWritesScriptAndSettings(t *testing.T) {
	root := t.TempDir()

	changed, err := InstallAntigravity(root)
	if err != nil {
		t.Fatalf("InstallAntigravity() error = %v", err)
	}
	if !changed {
		t.Fatalf("InstallAntigravity() changed = false, want true on first run")
	}

	scriptPath := filepath.Join(root, ".agents", "hooks", preToolUseScript)
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat %s: %v", scriptPath, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("%s is not executable: mode = %v", scriptPath, info.Mode())
	}

	group := readAntigravityHookGroup(t, filepath.Join(root, ".agents", "hooks.json"))
	if len(group["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d, want 1", len(group["PreToolUse"]))
	}
	wantMatcher := "write_to_file|replace_file_content|multi_replace_file_content"
	if group["PreToolUse"][0].Matcher != wantMatcher {
		t.Fatalf("PreToolUse matcher = %q, want %q", group["PreToolUse"][0].Matcher, wantMatcher)
	}
	if len(group["PreToolUse"][0].Hooks) != 1 || group["PreToolUse"][0].Hooks[0].Command != scriptPath {
		t.Fatalf("PreToolUse hook entry = %+v, want command %q", group["PreToolUse"][0].Hooks, scriptPath)
	}
}

func TestInstallAntigravityIsIdempotent(t *testing.T) {
	root := t.TempDir()

	if _, err := InstallAntigravity(root); err != nil {
		t.Fatalf("first InstallAntigravity() error = %v", err)
	}

	changed, err := InstallAntigravity(root)
	if err != nil {
		t.Fatalf("second InstallAntigravity() error = %v", err)
	}
	if changed {
		t.Fatalf("second InstallAntigravity() changed = true, want false")
	}

	group := readAntigravityHookGroup(t, filepath.Join(root, ".agents", "hooks.json"))
	if len(group["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d after second install, want 1 (no duplicate)", len(group["PreToolUse"]))
	}
}

func TestInstallAntigravityPreservesOtherNamedHookGroups(t *testing.T) {
	root := t.TempDir()

	agentsDir := filepath.Join(root, ".agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Matches the shape of Antigravity's own docs example: an unrelated
	// named hook group at the top level, sibling to whatever Loom owns.
	existing := `{
		"my-linter-hook": {
			"PostToolUse": [
				{"matcher": "run_command", "hooks": [{"type": "command", "command": "./scripts/lint.sh", "timeout": 10}]}
			]
		},
		"safety-gate": {
			"enabled": false,
			"PreToolUse": [
				{"matcher": "run_command", "hooks": [{"command": "./scripts/safety-check.sh"}]}
			]
		}
	}`
	settingsPath := filepath.Join(agentsDir, "hooks.json")
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("writing existing hooks.json: %v", err)
	}

	if _, err := InstallAntigravity(root); err != nil {
		t.Fatalf("InstallAntigravity() error = %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading hooks.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing hooks.json: %v", err)
	}

	if _, ok := raw["my-linter-hook"]; !ok {
		t.Fatalf("my-linter-hook key was not preserved")
	}
	if _, ok := raw["safety-gate"]; !ok {
		t.Fatalf("safety-gate key was not preserved")
	}

	var linterGroup map[string][]hookMatcher
	if err := json.Unmarshal(raw["my-linter-hook"], &linterGroup); err != nil {
		t.Fatalf("parsing my-linter-hook: %v", err)
	}
	if len(linterGroup["PostToolUse"]) != 1 || linterGroup["PostToolUse"][0].Hooks[0].Command != "./scripts/lint.sh" {
		t.Fatalf("my-linter-hook contents changed: %+v", linterGroup)
	}

	group := readAntigravityHookGroup(t, settingsPath)
	if len(group["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d, want 1", len(group["PreToolUse"]))
	}
}

func readAntigravityHookGroup(t *testing.T, path string) map[string][]hookMatcher {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	var group map[string][]hookMatcher
	if err := json.Unmarshal(raw[antigravityHookName], &group); err != nil {
		t.Fatalf("parsing %s in %s: %v", antigravityHookName, path, err)
	}

	return group
}

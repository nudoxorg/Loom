package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCodexWritesScriptsAndSettings(t *testing.T) {
	root := t.TempDir()

	changed, err := InstallCodex(root)
	if err != nil {
		t.Fatalf("InstallCodex() error = %v", err)
	}
	if !changed {
		t.Fatalf("InstallCodex() changed = false, want true on first run")
	}

	for _, name := range []string{sessionStartScript, preToolUseScript} {
		path := filepath.Join(root, ".codex", "hooks", name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable: mode = %v", path, info.Mode())
		}
	}

	settings := readHooksFile(t, filepath.Join(root, ".codex", "hooks.json"))
	if len(settings["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d, want 1", len(settings["PreToolUse"]))
	}
	if settings["PreToolUse"][0].Matcher != codexPreToolUseMatcher {
		t.Fatalf("PreToolUse matcher = %q, want %q", settings["PreToolUse"][0].Matcher, codexPreToolUseMatcher)
	}
	if len(settings["SessionStart"]) != 1 {
		t.Fatalf("SessionStart groups = %d, want 1", len(settings["SessionStart"]))
	}
}

func TestInstallCodexIsIdempotent(t *testing.T) {
	root := t.TempDir()

	if _, err := InstallCodex(root); err != nil {
		t.Fatalf("first InstallCodex() error = %v", err)
	}

	changed, err := InstallCodex(root)
	if err != nil {
		t.Fatalf("second InstallCodex() error = %v", err)
	}
	if changed {
		t.Fatalf("second InstallCodex() changed = true, want false")
	}

	settings := readHooksFile(t, filepath.Join(root, ".codex", "hooks.json"))
	if len(settings["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d after second install, want 1 (no duplicate)", len(settings["PreToolUse"]))
	}
}

func TestInstallCodexPreservesExistingSettings(t *testing.T) {
	root := t.TempDir()

	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// A hand-written entry using fields hookEntry doesn't itself set
	// (statusMessage, timeout) — these must survive a Loom install
	// untouched, since mergeSettingsFile round-trips the whole file.
	existing := `{
		"description": "Team lifecycle hooks",
		"hooks": {
			"Stop": [
				{"hooks": [{"type": "command", "command": "some-other-tool", "statusMessage": "Running policy check", "timeout": 30}]}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(codexDir, "hooks.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("writing existing hooks.json: %v", err)
	}

	if _, err := InstallCodex(root); err != nil {
		t.Fatalf("InstallCodex() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(codexDir, "hooks.json"))
	if err != nil {
		t.Fatalf("reading hooks.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing hooks.json: %v", err)
	}

	var description string
	if err := json.Unmarshal(raw["description"], &description); err != nil {
		t.Fatalf("parsing description: %v", err)
	}
	if description != "Team lifecycle hooks" {
		t.Fatalf("description = %q, want %q (must survive install)", description, "Team lifecycle hooks")
	}

	settings := readHooksFile(t, filepath.Join(codexDir, "hooks.json"))
	if len(settings["Stop"]) != 1 || settings["Stop"][0].Hooks[0].Command != "some-other-tool" {
		t.Fatalf("existing Stop hook was not preserved: %+v", settings["Stop"])
	}
	if len(settings["PreToolUse"]) != 1 {
		t.Fatalf("PreToolUse groups = %d, want 1", len(settings["PreToolUse"]))
	}

	// The unknown fields on the pre-existing entry (statusMessage, timeout)
	// must round-trip byte-for-byte, not just the fields hookEntry models.
	stopEntryJSON, err := json.Marshal(settings["Stop"][0].Hooks[0])
	if err != nil {
		t.Fatalf("marshaling preserved Stop entry: %v", err)
	}
	var stopEntry map[string]any
	if err := json.Unmarshal(stopEntryJSON, &stopEntry); err != nil {
		t.Fatalf("unmarshaling preserved Stop entry: %v", err)
	}
	if stopEntry["statusMessage"] != "Running policy check" {
		t.Fatalf("statusMessage = %v, want %q (must survive install)", stopEntry["statusMessage"], "Running policy check")
	}
	if stopEntry["timeout"] != float64(30) {
		t.Fatalf("timeout = %v, want 30 (must survive install)", stopEntry["timeout"])
	}
}

func TestInstallCodexGlobalWritesScriptAndSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	changed, err := InstallCodexGlobal()
	if err != nil {
		t.Fatalf("InstallCodexGlobal() error = %v", err)
	}
	if !changed {
		t.Fatalf("InstallCodexGlobal() changed = false, want true on first run")
	}

	path := filepath.Join(home, ".codex", "hooks", globalRemindScript)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("%s is not executable: mode = %v", path, info.Mode())
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if !strings.Contains(string(content), "git rev-parse --is-inside-work-tree") {
		t.Fatalf("global script does not gate on being inside a git repo:\n%s", content)
	}

	settings := readHooksFile(t, filepath.Join(home, ".codex", "hooks.json"))
	if len(settings["SessionStart"]) != 1 {
		t.Fatalf("SessionStart groups = %d, want 1", len(settings["SessionStart"]))
	}
	if settings["SessionStart"][0].Matcher != "" {
		t.Fatalf("SessionStart matcher = %q, want empty (not tool-scoped)", settings["SessionStart"][0].Matcher)
	}
}

func TestInstallCodexGlobalIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := InstallCodexGlobal(); err != nil {
		t.Fatalf("first InstallCodexGlobal() error = %v", err)
	}

	changed, err := InstallCodexGlobal()
	if err != nil {
		t.Fatalf("second InstallCodexGlobal() error = %v", err)
	}
	if changed {
		t.Fatalf("second InstallCodexGlobal() changed = true, want false")
	}

	settings := readHooksFile(t, filepath.Join(home, ".codex", "hooks.json"))
	if len(settings["SessionStart"]) != 1 {
		t.Fatalf("SessionStart groups = %d after second install, want 1 (no duplicate)", len(settings["SessionStart"]))
	}
}

func TestInstallCodexAndInstallCodexGlobalDoNotInterfere(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	project := t.TempDir()

	if _, err := InstallCodex(project); err != nil {
		t.Fatalf("InstallCodex() error = %v", err)
	}
	if _, err := InstallCodexGlobal(); err != nil {
		t.Fatalf("InstallCodexGlobal() error = %v", err)
	}

	projectSettings := readHooksFile(t, filepath.Join(project, ".codex", "hooks.json"))
	if len(projectSettings["PreToolUse"]) != 1 {
		t.Fatalf("project PreToolUse groups = %d, want 1", len(projectSettings["PreToolUse"]))
	}

	globalSettings := readHooksFile(t, filepath.Join(home, ".codex", "hooks.json"))
	if len(globalSettings["PreToolUse"]) != 0 {
		t.Fatalf("global settings picked up a PreToolUse hook, want none: %+v", globalSettings["PreToolUse"])
	}
	if len(globalSettings["SessionStart"]) != 1 {
		t.Fatalf("global SessionStart groups = %d, want 1", len(globalSettings["SessionStart"]))
	}
}

func TestInstallClaudeAndInstallCodexDoNotInterfere(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	project := t.TempDir()

	if _, err := Install(project); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := InstallGlobal(); err != nil {
		t.Fatalf("InstallGlobal() error = %v", err)
	}
	if _, err := InstallCodex(project); err != nil {
		t.Fatalf("InstallCodex() error = %v", err)
	}
	if _, err := InstallCodexGlobal(); err != nil {
		t.Fatalf("InstallCodexGlobal() error = %v", err)
	}

	claudeProjectSettings := readHooksFile(t, filepath.Join(project, ".claude", "settings.json"))
	if len(claudeProjectSettings["PreToolUse"]) != 1 {
		t.Fatalf("Claude project PreToolUse groups = %d, want 1", len(claudeProjectSettings["PreToolUse"]))
	}
	if claudeProjectSettings["PreToolUse"][0].Matcher != "Write|Edit|MultiEdit" {
		t.Fatalf("Claude PreToolUse matcher = %q, want Write|Edit|MultiEdit", claudeProjectSettings["PreToolUse"][0].Matcher)
	}

	codexProjectSettings := readHooksFile(t, filepath.Join(project, ".codex", "hooks.json"))
	if len(codexProjectSettings["PreToolUse"]) != 1 {
		t.Fatalf("Codex project PreToolUse groups = %d, want 1", len(codexProjectSettings["PreToolUse"]))
	}
	if codexProjectSettings["PreToolUse"][0].Matcher != codexPreToolUseMatcher {
		t.Fatalf("Codex PreToolUse matcher = %q, want %q", codexProjectSettings["PreToolUse"][0].Matcher, codexPreToolUseMatcher)
	}
}

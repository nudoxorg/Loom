package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCursorWritesScriptsAndSettings(t *testing.T) {
	root := t.TempDir()

	changed, err := InstallCursor(root)
	if err != nil {
		t.Fatalf("InstallCursor() error = %v", err)
	}
	if !changed {
		t.Fatalf("InstallCursor() changed = false, want true on first run")
	}

	for _, name := range []string{sessionStartScript, cursorPostToolUseScript} {
		path := filepath.Join(root, ".cursor", "hooks", name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable: mode = %v", path, info.Mode())
		}
	}

	settingsPath := filepath.Join(root, ".cursor", "hooks.json")
	settings := readCursorHooksFile(t, settingsPath)

	if len(settings["postToolUse"]) != 1 {
		t.Fatalf("postToolUse handlers = %d, want 1", len(settings["postToolUse"]))
	}
	if settings["postToolUse"][0].Matcher != "Write" {
		t.Fatalf("postToolUse matcher = %q, want Write", settings["postToolUse"][0].Matcher)
	}
	if len(settings["sessionStart"]) != 1 {
		t.Fatalf("sessionStart handlers = %d, want 1", len(settings["sessionStart"]))
	}
	if settings["sessionStart"][0].Matcher != "" {
		t.Fatalf("sessionStart matcher = %q, want empty", settings["sessionStart"][0].Matcher)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading %s: %v", settingsPath, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing %s: %v", settingsPath, err)
	}
	var version int
	if err := json.Unmarshal(raw["version"], &version); err != nil {
		t.Fatalf("parsing version: %v", err)
	}
	if version != 1 {
		t.Fatalf("version = %d, want 1", version)
	}
}

func TestInstallCursorIsIdempotent(t *testing.T) {
	root := t.TempDir()

	if _, err := InstallCursor(root); err != nil {
		t.Fatalf("first InstallCursor() error = %v", err)
	}

	changed, err := InstallCursor(root)
	if err != nil {
		t.Fatalf("second InstallCursor() error = %v", err)
	}
	if changed {
		t.Fatalf("second InstallCursor() changed = true, want false")
	}

	settings := readCursorHooksFile(t, filepath.Join(root, ".cursor", "hooks.json"))
	if len(settings["postToolUse"]) != 1 {
		t.Fatalf("postToolUse handlers = %d after second install, want 1 (no duplicate)", len(settings["postToolUse"]))
	}
}

func TestInstallCursorPreservesExistingSettings(t *testing.T) {
	root := t.TempDir()

	cursorDir := filepath.Join(root, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// A hand-written entry using fields cursorHookHandler doesn't itself
	// set (loop_limit, failClosed) — these must survive a Loom install
	// untouched. No "version" key here either, to confirm Loom doesn't
	// invent one for a file that already exists without it.
	existing := `{
		"hooks": {
			"stop": [
				{"command": "./hooks/audit.sh", "loop_limit": 10, "failClosed": true}
			]
		}
	}`
	settingsPath := filepath.Join(cursorDir, "hooks.json")
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("writing existing hooks.json: %v", err)
	}

	if _, err := InstallCursor(root); err != nil {
		t.Fatalf("InstallCursor() error = %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading hooks.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing hooks.json: %v", err)
	}
	if _, ok := raw["version"]; ok {
		t.Fatalf("version key was invented for a pre-existing file that never had one: %s", raw["version"])
	}

	settings := readCursorHooksFile(t, settingsPath)
	if len(settings["stop"]) != 1 || settings["stop"][0].Command != "./hooks/audit.sh" {
		t.Fatalf("existing stop hook was not preserved: %+v", settings["stop"])
	}
	if len(settings["postToolUse"]) != 1 {
		t.Fatalf("postToolUse handlers = %d, want 1", len(settings["postToolUse"]))
	}

	stopEntryJSON, err := json.Marshal(settings["stop"][0])
	if err != nil {
		t.Fatalf("marshaling preserved stop entry: %v", err)
	}
	var stopEntry map[string]any
	if err := json.Unmarshal(stopEntryJSON, &stopEntry); err != nil {
		t.Fatalf("unmarshaling preserved stop entry: %v", err)
	}
	if stopEntry["loop_limit"] != float64(10) {
		t.Fatalf("loop_limit = %v, want 10 (must survive install)", stopEntry["loop_limit"])
	}
	if stopEntry["failClosed"] != true {
		t.Fatalf("failClosed = %v, want true (must survive install)", stopEntry["failClosed"])
	}
}

func TestInstallCursorGlobalWritesScriptAndSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	changed, err := InstallCursorGlobal()
	if err != nil {
		t.Fatalf("InstallCursorGlobal() error = %v", err)
	}
	if !changed {
		t.Fatalf("InstallCursorGlobal() changed = false, want true on first run")
	}

	path := filepath.Join(home, ".cursor", "hooks", globalRemindScript)
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
	if strings.Contains(string(content), "git rev-parse") {
		t.Fatalf("global script should fire unconditionally, not gate on being inside a git repo:\n%s", content)
	}

	settings := readCursorHooksFile(t, filepath.Join(home, ".cursor", "hooks.json"))
	if len(settings["sessionStart"]) != 1 {
		t.Fatalf("sessionStart handlers = %d, want 1", len(settings["sessionStart"]))
	}
	if !filepath.IsAbs(settings["sessionStart"][0].Command) {
		t.Fatalf("global sessionStart command = %q, want an absolute path", settings["sessionStart"][0].Command)
	}
}

func TestInstallCursorGlobalIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := InstallCursorGlobal(); err != nil {
		t.Fatalf("first InstallCursorGlobal() error = %v", err)
	}

	changed, err := InstallCursorGlobal()
	if err != nil {
		t.Fatalf("second InstallCursorGlobal() error = %v", err)
	}
	if changed {
		t.Fatalf("second InstallCursorGlobal() changed = true, want false")
	}

	settings := readCursorHooksFile(t, filepath.Join(home, ".cursor", "hooks.json"))
	if len(settings["sessionStart"]) != 1 {
		t.Fatalf("sessionStart handlers = %d after second install, want 1 (no duplicate)", len(settings["sessionStart"]))
	}
}

func TestInstallCursorAndInstallCursorGlobalDoNotInterfere(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	project := t.TempDir()

	if _, err := InstallCursor(project); err != nil {
		t.Fatalf("InstallCursor() error = %v", err)
	}
	if _, err := InstallCursorGlobal(); err != nil {
		t.Fatalf("InstallCursorGlobal() error = %v", err)
	}

	projectSettings := readCursorHooksFile(t, filepath.Join(project, ".cursor", "hooks.json"))
	if len(projectSettings["postToolUse"]) != 1 {
		t.Fatalf("project postToolUse handlers = %d, want 1", len(projectSettings["postToolUse"]))
	}

	globalSettings := readCursorHooksFile(t, filepath.Join(home, ".cursor", "hooks.json"))
	if len(globalSettings["postToolUse"]) != 0 {
		t.Fatalf("global settings picked up a postToolUse hook, want none: %+v", globalSettings["postToolUse"])
	}
	if len(globalSettings["sessionStart"]) != 1 {
		t.Fatalf("global sessionStart handlers = %d, want 1", len(globalSettings["sessionStart"]))
	}
}

func TestInstallClaudeCodexAndCursorDoNotInterfere(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	project := t.TempDir()

	if _, err := Install(project); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := InstallCodex(project); err != nil {
		t.Fatalf("InstallCodex() error = %v", err)
	}
	if _, err := InstallCursor(project); err != nil {
		t.Fatalf("InstallCursor() error = %v", err)
	}

	claudeSettings := readHooksFile(t, filepath.Join(project, ".claude", "settings.json"))
	if len(claudeSettings["PreToolUse"]) != 1 {
		t.Fatalf("Claude PreToolUse groups = %d, want 1", len(claudeSettings["PreToolUse"]))
	}

	codexSettings := readHooksFile(t, filepath.Join(project, ".codex", "hooks.json"))
	if len(codexSettings["PreToolUse"]) != 1 {
		t.Fatalf("Codex PreToolUse groups = %d, want 1", len(codexSettings["PreToolUse"]))
	}

	cursorSettings := readCursorHooksFile(t, filepath.Join(project, ".cursor", "hooks.json"))
	if len(cursorSettings["postToolUse"]) != 1 {
		t.Fatalf("Cursor postToolUse handlers = %d, want 1", len(cursorSettings["postToolUse"]))
	}
}

func readCursorHooksFile(t *testing.T, path string) map[string][]cursorHookHandler {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	var hooksByEvent map[string][]cursorHookHandler
	if err := json.Unmarshal(raw["hooks"], &hooksByEvent); err != nil {
		t.Fatalf("parsing hooks in %s: %v", path, err)
	}

	return hooksByEvent
}

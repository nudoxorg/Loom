package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallClaudeWritesGlobalSessionAndPromptHooks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	changed, err := Install(HarnessClaude)
	if err != nil {
		t.Fatalf("Install(claude) error = %v", err)
	}
	if !changed {
		t.Fatal("Install(claude) changed = false, want true")
	}

	hooksDir := filepath.Join(home, ".claude", "hooks")
	if content := requireExecutable(t, filepath.Join(hooksDir, sessionStartScript)); !strings.Contains(content, `"hookEventName":"SessionStart"`) {
		t.Fatalf("session hook does not emit SessionStart payload:\n%s", content)
	}
	if content := requireExecutable(t, filepath.Join(hooksDir, promptSubmitScript)); !strings.Contains(content, `"hookEventName":"UserPromptSubmit"`) {
		t.Fatalf("prompt hook does not emit UserPromptSubmit payload:\n%s", content)
	}

	hooksByEvent := readNestedHooks(t, filepath.Join(home, ".claude", "settings.json"))
	if len(hooksByEvent["SessionStart"]) != 1 || len(hooksByEvent["UserPromptSubmit"]) != 1 {
		t.Fatalf("Claude hooks = %+v, want one SessionStart and one UserPromptSubmit", hooksByEvent)
	}
	if len(hooksByEvent["PreToolUse"]) != 0 {
		t.Fatalf("Claude installed project-style PreToolUse hook: %+v", hooksByEvent["PreToolUse"])
	}

	changed, err = Install(HarnessClaude)
	if err != nil {
		t.Fatalf("second Install(claude) error = %v", err)
	}
	if changed {
		t.Fatal("second Install(claude) changed = true, want false")
	}
}

func TestInstallClaudePreservesExistingSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "model": "opus",
  "hooks": {
    "Stop": [{"hooks": [{"type": "command", "command": "other-tool", "timeout": 30}]}]
  }
}`
	path := filepath.Join(configDir, "settings.json")
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(HarnessClaude); err != nil {
		t.Fatalf("Install(claude) error = %v", err)
	}
	raw := readRawObject(t, path)
	var model string
	if err := json.Unmarshal(raw["model"], &model); err != nil || model != "opus" {
		t.Fatalf("model not preserved: value=%q error=%v", model, err)
	}
	hooksByEvent := readNestedHooks(t, path)
	preserved, err := json.Marshal(hooksByEvent["Stop"][0].Hooks[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(preserved), `"timeout":30`) {
		t.Fatalf("unknown handler field was not preserved: %s", preserved)
	}
}

func TestInstallClaudeUpgradesOldGlobalHookWithoutDuplicate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hooksDir := filepath.Join(home, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sessionPath := filepath.Join(hooksDir, sessionStartScript)
	if err := os.WriteFile(sessionPath, []byte("#!/bin/bash\necho old reminder\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":` + quotedJSON(t, sessionPath) + `}]}]}}`
	if err := os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(HarnessClaude); err != nil {
		t.Fatalf("Install(claude) error = %v", err)
	}
	hooksByEvent := readNestedHooks(t, filepath.Join(home, ".claude", "settings.json"))
	if len(hooksByEvent["SessionStart"]) != 1 {
		t.Fatalf("SessionStart groups = %d, want old global entry upgraded in place without duplication", len(hooksByEvent["SessionStart"]))
	}
	if len(hooksByEvent["UserPromptSubmit"]) != 1 {
		t.Fatalf("UserPromptSubmit groups = %d, want newly added prompt reminder", len(hooksByEvent["UserPromptSubmit"]))
	}
}

func quotedJSON(t *testing.T, value string) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

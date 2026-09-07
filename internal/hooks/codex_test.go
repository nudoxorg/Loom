package hooks

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCodexWritesGlobalSessionAndPromptHooks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	changed, err := Install(HarnessCodex)
	if err != nil {
		t.Fatalf("Install(codex) error = %v", err)
	}
	if !changed {
		t.Fatal("Install(codex) changed = false, want true")
	}

	hooksDir := filepath.Join(home, ".codex", "hooks")
	if content := requireExecutable(t, filepath.Join(hooksDir, sessionStartScript)); !strings.Contains(content, `"hookEventName":"SessionStart"`) {
		t.Fatalf("session hook does not emit SessionStart payload:\n%s", content)
	}
	if content := requireExecutable(t, filepath.Join(hooksDir, promptSubmitScript)); !strings.Contains(content, `"hookEventName":"UserPromptSubmit"`) {
		t.Fatalf("prompt hook does not emit UserPromptSubmit payload:\n%s", content)
	}

	hooksByEvent := readNestedHooks(t, filepath.Join(home, ".codex", "hooks.json"))
	if len(hooksByEvent["SessionStart"]) != 1 || len(hooksByEvent["UserPromptSubmit"]) != 1 {
		t.Fatalf("Codex hooks = %+v, want one SessionStart and one UserPromptSubmit", hooksByEvent)
	}
	if len(hooksByEvent["PreToolUse"]) != 0 {
		t.Fatalf("Codex installed project-style PreToolUse hook: %+v", hooksByEvent["PreToolUse"])
	}

	changed, err = Install(HarnessCodex)
	if err != nil {
		t.Fatalf("second Install(codex) error = %v", err)
	}
	if changed {
		t.Fatal("second Install(codex) changed = true, want false")
	}
}

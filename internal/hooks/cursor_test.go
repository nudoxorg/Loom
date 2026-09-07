package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCursorWritesOnlySupportedGlobalReminder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	changed, err := Install(HarnessCursor)
	if err != nil {
		t.Fatalf("Install(cursor) error = %v", err)
	}
	if !changed {
		t.Fatal("Install(cursor) changed = false, want true")
	}

	scriptPath := filepath.Join(home, ".cursor", "hooks", sessionStartScript)
	if content := requireExecutable(t, scriptPath); !strings.Contains(content, `"additional_context"`) {
		t.Fatalf("Cursor hook does not emit additional_context payload:\n%s", content)
	}

	configPath := filepath.Join(home, ".cursor", "hooks.json")
	raw := readRawObject(t, configPath)
	if string(raw["version"]) != "1" {
		t.Fatalf("Cursor version = %s, want 1", raw["version"])
	}
	hooksByEvent := readCursorHooks(t, configPath)
	if len(hooksByEvent["sessionStart"]) != 1 {
		t.Fatalf("sessionStart hooks = %+v, want one", hooksByEvent["sessionStart"])
	}
	if len(hooksByEvent["beforeSubmitPrompt"]) != 0 || len(hooksByEvent["postToolUse"]) != 0 {
		t.Fatalf("Cursor installed an unsupported prompt/edit reminder: %+v", hooksByEvent)
	}

	changed, err = Install(HarnessCursor)
	if err != nil {
		t.Fatalf("second Install(cursor) error = %v", err)
	}
	if changed {
		t.Fatal("second Install(cursor) changed = true, want false")
	}
}

func TestInstallCursorPreservesExistingConfigAndHandlerFields(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "version": 7,
  "label": "personal hooks",
  "hooks": {
    "stop": [{"command": "other-tool", "timeout": 30}]
  }
}`
	path := filepath.Join(configDir, "hooks.json")
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(HarnessCursor); err != nil {
		t.Fatalf("Install(cursor) error = %v", err)
	}
	raw := readRawObject(t, path)
	if string(raw["version"]) != "7" || string(raw["label"]) != `"personal hooks"` {
		t.Fatalf("top-level config not preserved: version=%s label=%s", raw["version"], raw["label"])
	}
	hooksByEvent := readCursorHooks(t, path)
	preserved, err := json.Marshal(hooksByEvent["stop"][0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(preserved), `"timeout":30`) {
		t.Fatalf("unknown handler field was not preserved: %s", preserved)
	}
}

package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseHarness(t *testing.T) {
	tests := map[string]Harness{
		"claude":      HarnessClaude,
		"Claude-Code": HarnessClaude,
		"codex":       HarnessCodex,
		" cursor ":    HarnessCursor,
	}
	for input, want := range tests {
		got, err := ParseHarness(input)
		if err != nil {
			t.Fatalf("ParseHarness(%q) error = %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseHarness(%q) = %q, want %q", input, got, want)
		}
	}
	if _, err := ParseHarness("antigravity"); err == nil {
		t.Fatal("ParseHarness(antigravity) error = nil, want unsupported harness error")
	}
}

func TestInstallAllWritesOnlyGlobalConfigs(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)

	changed, err := InstallAll()
	if err != nil {
		t.Fatalf("InstallAll() error = %v", err)
	}
	if !changed {
		t.Fatal("InstallAll() changed = false, want true on first run")
	}

	for _, path := range []string{
		filepath.Join(home, ".claude", "settings.json"),
		filepath.Join(home, ".codex", "hooks.json"),
		filepath.Join(home, ".cursor", "hooks.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
	}

	changed, err = InstallAll()
	if err != nil {
		t.Fatalf("second InstallAll() error = %v", err)
	}
	if changed {
		t.Fatal("second InstallAll() changed = true, want false")
	}
}

func readNestedHooks(t *testing.T, path string) map[string][]hookMatcher {
	t.Helper()

	raw := readRawObject(t, path)
	var hooksByEvent map[string][]hookMatcher
	if err := json.Unmarshal(raw["hooks"], &hooksByEvent); err != nil {
		t.Fatalf("parsing hooks in %s: %v", path, err)
	}
	return hooksByEvent
}

func readCursorHooks(t *testing.T, path string) map[string][]cursorHookHandler {
	t.Helper()

	raw := readRawObject(t, path)
	var hooksByEvent map[string][]cursorHookHandler
	if err := json.Unmarshal(raw["hooks"], &hooksByEvent); err != nil {
		t.Fatalf("parsing hooks in %s: %v", path, err)
	}
	return hooksByEvent
}

func readRawObject(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return raw
}

func requireExecutable(t *testing.T, path string) string {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("%s is not executable: mode = %v", path, info.Mode())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

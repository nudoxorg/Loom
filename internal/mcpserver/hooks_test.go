package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleHooksInstallRequiresHarness(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	res, err := handleHooksInstall(context.Background(), callWith(nil))
	if err != nil {
		t.Fatalf("handleHooksInstall() error = %v", err)
	}
	if !res.IsError {
		t.Fatalf("handleHooksInstall() without harness should error, got %q", resultText(t, res))
	}
}

func TestHandleHooksInstallRejectsUnknownHarness(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	res, err := handleHooksInstall(context.Background(), callWith(map[string]any{"harness": "antigravity"}))
	if err != nil {
		t.Fatalf("handleHooksInstall() error = %v", err)
	}
	if !res.IsError || !strings.Contains(resultText(t, res), "unsupported harness") {
		t.Fatalf("handleHooksInstall(antigravity) = %#v, want unsupported harness error", res)
	}
}

func TestHandleHooksInstallAllUsesGlobalHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	res, err := handleHooksInstall(context.Background(), callWith(map[string]any{"harness": "all"}))
	if err != nil {
		t.Fatalf("handleHooksInstall() error = %v", err)
	}
	if res.IsError {
		t.Fatalf("handleHooksInstall(all) returned error: %q", resultText(t, res))
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
}

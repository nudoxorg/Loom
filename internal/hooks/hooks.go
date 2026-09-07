// Package hooks installs global coding-agent hooks that remind agents to use
// Loom. Hooks only inject static context; they never invoke Loom themselves.
package hooks

import (
	"fmt"
	"strings"
)

// Harness identifies an agent harness whose global hook configuration Loom
// knows how to update.
type Harness string

const (
	HarnessClaude Harness = "claude"
	HarnessCodex  Harness = "codex"
	HarnessCursor Harness = "cursor"
)

var supportedHarnesses = [...]Harness{HarnessClaude, HarnessCodex, HarnessCursor}

// ParseHarness converts a user-facing harness name to its canonical value.
func ParseHarness(value string) (Harness, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "claude", "claude-code", "claude_code":
		return HarnessClaude, nil
	case "codex":
		return HarnessCodex, nil
	case "cursor":
		return HarnessCursor, nil
	default:
		return "", fmt.Errorf("unsupported harness %q (expected claude, codex, or cursor)", value)
	}
}

// Install updates one harness's global user configuration. It is idempotent
// and preserves configuration and hooks not owned by Loom.
func Install(harness Harness) (bool, error) {
	switch harness {
	case HarnessClaude:
		return installClaude()
	case HarnessCodex:
		return installCodex()
	case HarnessCursor:
		return installCursor()
	default:
		return false, fmt.Errorf("unsupported harness %q", harness)
	}
}

// InstallAll installs global hooks for every supported harness.
func InstallAll() (bool, error) {
	changed := false
	for _, harness := range supportedHarnesses {
		harnessChanged, err := Install(harness)
		if err != nil {
			return changed, fmt.Errorf("installing %s hooks: %w", harness, err)
		}
		changed = harnessChanged || changed
	}
	return changed, nil
}

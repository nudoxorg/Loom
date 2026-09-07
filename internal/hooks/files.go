package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func homeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return home, nil
}

func writeScriptIfChanged(path, content string) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil && string(existing) == content {
		if info, statErr := os.Stat(path); statErr == nil && info.Mode()&0o111 != 0 {
			return false, nil
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("creating hooks directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		return false, fmt.Errorf("writing %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return false, fmt.Errorf("making %s executable: %w", path, err)
	}
	return true, nil
}

// mergeNestedHooksFile updates Claude/Codex's shared nested hook shape.
func mergeNestedHooksFile(path string, wirings []hookWiring) (bool, error) {
	raw, err := readJSONObject(path)
	if err != nil {
		return false, err
	}

	hooksByEvent := map[string][]hookMatcher{}
	if existing, ok := raw["hooks"]; ok {
		if err := json.Unmarshal(existing, &hooksByEvent); err != nil {
			return false, fmt.Errorf("parsing hooks in %s: %w", path, err)
		}
	}

	changed := false
	for _, wiring := range wirings {
		changed = addNestedHookIfMissing(hooksByEvent, wiring) || changed
	}
	if !changed {
		return false, nil
	}

	hooksJSON, err := json.Marshal(hooksByEvent)
	if err != nil {
		return false, fmt.Errorf("encoding hooks: %w", err)
	}
	raw["hooks"] = hooksJSON
	return true, writeJSONObject(path, raw)
}

func addNestedHookIfMissing(hooksByEvent map[string][]hookMatcher, wiring hookWiring) bool {
	for _, group := range hooksByEvent[wiring.event] {
		for _, handler := range group.Hooks {
			if handler.Command == wiring.command {
				return false
			}
		}
	}
	hooksByEvent[wiring.event] = append(hooksByEvent[wiring.event], hookMatcher{
		Matcher: wiring.matcher,
		Hooks:   []hookEntry{{Type: "command", Command: wiring.command}},
	})
	return true
}

func mergeCursorHooksFile(path string, wirings []hookWiring) (bool, error) {
	raw, err := readJSONObject(path)
	if err != nil {
		return false, err
	}

	hooksByEvent := map[string][]cursorHookHandler{}
	if existing, ok := raw["hooks"]; ok {
		if err := json.Unmarshal(existing, &hooksByEvent); err != nil {
			return false, fmt.Errorf("parsing hooks in %s: %w", path, err)
		}
	}

	changed := false
	for _, wiring := range wirings {
		changed = addCursorHookIfMissing(hooksByEvent, wiring) || changed
	}
	if _, exists := raw["version"]; !exists {
		raw["version"] = json.RawMessage("1")
		changed = true
	}
	if !changed {
		return false, nil
	}

	hooksJSON, err := json.Marshal(hooksByEvent)
	if err != nil {
		return false, fmt.Errorf("encoding hooks: %w", err)
	}
	raw["hooks"] = hooksJSON
	return true, writeJSONObject(path, raw)
}

func addCursorHookIfMissing(hooksByEvent map[string][]cursorHookHandler, wiring hookWiring) bool {
	for _, handler := range hooksByEvent[wiring.event] {
		if handler.Command == wiring.command {
			return false
		}
	}
	hooksByEvent[wiring.event] = append(hooksByEvent[wiring.event], cursorHookHandler{
		Command: wiring.command,
		Matcher: wiring.matcher,
	})
	return true
}

func readJSONObject(path string) (map[string]json.RawMessage, error) {
	raw := map[string]json.RawMessage{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
	case os.IsNotExist(err):
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("creating hooks config directory: %w", err)
		}
	default:
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return raw, nil
}

func writeJSONObject(path string, raw map[string]json.RawMessage) error {
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

package hooks

import "path/filepath"

func installCursor() (bool, error) {
	home, err := homeDirectory()
	if err != nil {
		return false, err
	}

	scriptPath := filepath.Join(home, ".cursor", "hooks", sessionStartScript)
	scriptChanged, err := writeScriptIfChanged(scriptPath, cursorScriptContent(sessionStartMessage))
	if err != nil {
		return false, err
	}

	configChanged, err := mergeCursorHooksFile(filepath.Join(home, ".cursor", "hooks.json"), []hookWiring{
		{event: "sessionStart", command: scriptPath},
	})
	if err != nil {
		return false, err
	}
	return scriptChanged || configChanged, nil
}

package hooks

import "path/filepath"

func installCodex() (bool, error) {
	home, err := homeDirectory()
	if err != nil {
		return false, err
	}

	hooksDir := filepath.Join(home, ".codex", "hooks")
	sessionPath := filepath.Join(hooksDir, sessionStartScript)
	promptPath := filepath.Join(hooksDir, promptSubmitScript)

	sessionChanged, err := writeScriptIfChanged(sessionPath, nestedScriptContent("SessionStart", sessionStartMessage, HarnessCodex))
	if err != nil {
		return false, err
	}
	promptChanged, err := writeScriptIfChanged(promptPath, nestedScriptContent("UserPromptSubmit", promptSubmitMessage, HarnessCodex))
	if err != nil {
		return false, err
	}

	configChanged, err := mergeNestedHooksFile(filepath.Join(home, ".codex", "hooks.json"), []hookWiring{
		{event: "SessionStart", command: sessionPath},
		{event: "UserPromptSubmit", command: promptPath},
	})
	if err != nil {
		return false, err
	}
	return sessionChanged || promptChanged || configChanged, nil
}

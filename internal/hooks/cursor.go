package hooks

import (
	"os"
	"path/filepath"
)

func installCursor() (bool, error) {
	home, err := homeDirectory()
	if err != nil {
		return false, err
	}

	hooksDir := filepath.Join(home, ".cursor", "hooks")
	sessionPath := filepath.Join(hooksDir, sessionStartScript)
	promptPath := filepath.Join(hooksDir, promptSubmitScript)
	subagentPath := filepath.Join(hooksDir, cursorTaskScript)

	sessionChanged, err := writeScriptIfChanged(sessionPath, cursorScriptContent(sessionStartMessage))
	if err != nil {
		return false, err
	}
	// Cursor maps Claude Code's UserPromptSubmit event to beforeSubmitPrompt and
	// accepts the nested hookSpecificOutput response format. Use that compatibility
	// path so the reminder is injected into agent context instead of shown only as
	// a native beforeSubmitPrompt user_message.
	promptChanged, err := writeScriptIfChanged(promptPath, nestedScriptContent("UserPromptSubmit", promptSubmitMessage, HarnessCursor))
	if err != nil {
		return false, err
	}
	executable, err := os.Executable()
	if err != nil {
		return false, err
	}
	subagentChanged, err := writeScriptIfChanged(subagentPath, cursorSubagentScriptContent(executable))
	if err != nil {
		return false, err
	}

	configChanged, err := mergeCursorHooksFile(filepath.Join(home, ".cursor", "hooks.json"), []hookWiring{
		{event: "sessionStart", command: sessionPath},
		{event: "beforeSubmitPrompt", command: promptPath},
		{event: "subagentStart", command: subagentPath},
		{event: "preToolUse", matcher: "^Task$", command: subagentPath},
	})
	if err != nil {
		return false, err
	}
	return sessionChanged || promptChanged || subagentChanged || configChanged, nil
}

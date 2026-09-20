package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const subagentContextCloseMarker = "</loom-subagent-context>"

type cursorHookInput struct {
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

type cursorHookOutput struct {
	Permission   string                     `json:"permission"`
	UpdatedInput map[string]json.RawMessage `json:"updated_input,omitempty"`
}

// HandleCursorSubagentHook handles both Cursor's subagentStart lifecycle hook
// and the Task-scoped preToolUse hook used to inject agent-visible context.
// Invalid or unfamiliar input is allowed through so a Loom reminder can never
// prevent Cursor from spawning a subagent.
func HandleCursorSubagentHook(in io.Reader, out, diagnostic io.Writer) error {
	var input cursorHookInput
	if err := json.NewDecoder(in).Decode(&input); err != nil {
		fmt.Fprintf(diagnostic, "loom: cursor subagent hook received invalid JSON: %v\n", err)
		return writeCursorHookOutput(out, cursorHookOutput{Permission: "allow"})
	}

	response := cursorHookOutput{Permission: "allow"}
	if input.HookEventName != "preToolUse" || input.ToolName != "Task" {
		return writeCursorHookOutput(out, response)
	}

	toolInput := map[string]json.RawMessage{}
	if err := json.Unmarshal(input.ToolInput, &toolInput); err != nil {
		fmt.Fprintf(diagnostic, "loom: cursor Task hook received invalid tool_input: %v\n", err)
		return writeCursorHookOutput(out, response)
	}
	if cursorTaskIsResume(toolInput) {
		return writeCursorHookOutput(out, response)
	}

	promptField, prompt, ok := cursorTaskPrompt(toolInput)
	if !ok || strings.Contains(prompt, subagentContextMarker) {
		return writeCursorHookOutput(out, response)
	}

	prompt = subagentContextMarker + "\n" + subagentStartMessage + "\n" + subagentContextCloseMarker + "\n\n" + prompt
	promptJSON, err := json.Marshal(prompt)
	if err != nil {
		fmt.Fprintf(diagnostic, "loom: cursor Task hook could not encode context: %v\n", err)
		return writeCursorHookOutput(out, response)
	}
	toolInput[promptField] = promptJSON
	response.UpdatedInput = toolInput
	return writeCursorHookOutput(out, response)
}

func cursorTaskIsResume(toolInput map[string]json.RawMessage) bool {
	raw, ok := toolInput["resume"]
	if !ok {
		return false
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || trimmed == "false" || trimmed == `""` {
		return false
	}
	return true
}

func cursorTaskPrompt(toolInput map[string]json.RawMessage) (string, string, bool) {
	for _, field := range []string{"prompt", "task"} {
		raw, ok := toolInput[field]
		if !ok {
			continue
		}
		var prompt string
		if json.Unmarshal(raw, &prompt) == nil {
			return field, prompt, true
		}
	}
	return "", "", false
}

func writeCursorHookOutput(out io.Writer, response cursorHookOutput) error {
	return json.NewEncoder(out).Encode(response)
}

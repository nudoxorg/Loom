package hooks

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestHandleCursorSubagentHookPrefixesTaskPromptAndPreservesInput(t *testing.T) {
	input := `{
  "hook_event_name": "preToolUse",
  "tool_name": "Task",
  "tool_input": {
    "description": "inspect hooks",
    "prompt": "Review the hook implementation.",
    "model": "fast",
    "custom": {"keep": true}
  }
}`

	output := runCursorSubagentHook(t, input)
	if output.Permission != "allow" {
		t.Fatalf("permission = %q, want allow", output.Permission)
	}
	if output.UpdatedInput == nil {
		t.Fatal("updated_input missing")
	}
	var prompt string
	if err := json.Unmarshal(output.UpdatedInput["prompt"], &prompt); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(prompt, subagentContextMarker+"\n"+subagentStartMessage) || !strings.HasSuffix(prompt, "Review the hook implementation.") {
		t.Fatalf("prompt was not prefixed correctly: %q", prompt)
	}
	var custom map[string]bool
	if err := json.Unmarshal(output.UpdatedInput["custom"], &custom); err != nil {
		t.Fatal(err)
	}
	if string(output.UpdatedInput["model"]) != `"fast"` || !custom["keep"] {
		t.Fatalf("unrelated Task input was not preserved: %+v", output.UpdatedInput)
	}
}

func TestHandleCursorSubagentHookDoesNotDoubleInjectOrChangeResume(t *testing.T) {
	tests := map[string]string{
		"marked": `{"hook_event_name":"preToolUse","tool_name":"Task","tool_input":{"prompt":"` + subagentContextMarker + ` already present"}}`,
		"resume": `{"hook_event_name":"preToolUse","tool_name":"Task","tool_input":{"prompt":"continue","resume":"agent-123"}}`,
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			output := runCursorSubagentHook(t, input)
			if output.Permission != "allow" || output.UpdatedInput != nil {
				t.Fatalf("output = %+v, want allow without updated_input", output)
			}
		})
	}
}

func TestHandleCursorSubagentHookSupportsTaskField(t *testing.T) {
	output := runCursorSubagentHook(t, `{"hook_event_name":"preToolUse","tool_name":"Task","tool_input":{"task":"Investigate the bug."}}`)
	var task string
	if err := json.Unmarshal(output.UpdatedInput["task"], &task); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(task, subagentStartMessage) {
		t.Fatalf("task = %q, want Loom context", task)
	}
}

func TestHandleCursorSubagentHookFailsOpen(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := HandleCursorSubagentHook(strings.NewReader("not-json"), &stdout, &stderr); err != nil {
		t.Fatalf("HandleCursorSubagentHook() error = %v", err)
	}
	var output cursorHookOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("invalid output %q: %v", stdout.String(), err)
	}
	if output.Permission != "allow" || !strings.Contains(stderr.String(), "invalid JSON") {
		t.Fatalf("stdout=%q stderr=%q, want fail-open response and diagnostic", stdout.String(), stderr.String())
	}
}

func TestHandleCursorSubagentStartAllowsLifecycleEvent(t *testing.T) {
	output := runCursorSubagentHook(t, `{"hook_event_name":"subagentStart","subagent_type":"explore"}`)
	if output.Permission != "allow" || output.UpdatedInput != nil {
		t.Fatalf("output = %+v, want allow without updated_input", output)
	}
}

func runCursorSubagentHook(t *testing.T, input string) cursorHookOutput {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if err := HandleCursorSubagentHook(strings.NewReader(input), &stdout, &stderr); err != nil {
		t.Fatalf("HandleCursorSubagentHook() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected diagnostic: %s", stderr.String())
	}
	var output cursorHookOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("invalid output %q: %v", stdout.String(), err)
	}
	return output
}

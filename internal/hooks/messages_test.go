package hooks

import "testing"

func TestExistingRootAgentMessagesRemainUnchanged(t *testing.T) {
	const wantSession = "Loom is available globally for coordination between coding agents. Prefer its MCP tools over the CLI. At the start of work, call loom_global_all for cross-project context. Before editing, claim the relevant path with loom_claim; release it when done; and record meaningful decisions with loom_log. Project-scoped tools require your actual current working directory as cwd."
	const wantPrompt = "Remember to use Loom's MCP tools for shared agent coordination: check loom_global_all when context may have changed, claim paths before editing, release them when done, and log meaningful decisions. Always pass your actual current working directory as cwd to project-scoped tools."

	if sessionStartMessage != wantSession {
		t.Fatalf("sessionStartMessage changed:\n got: %q\nwant: %q", sessionStartMessage, wantSession)
	}
	if promptSubmitMessage != wantPrompt {
		t.Fatalf("promptSubmitMessage changed:\n got: %q\nwant: %q", promptSubmitMessage, wantPrompt)
	}
}

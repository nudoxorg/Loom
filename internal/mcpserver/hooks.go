package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/nudoxorg/loom/internal/hooks"
)

func registerHooksTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool(
			"loom_hooks_install",
			mcp.WithDescription("Install global Loom reminder hooks for root agents and subagents in Claude Code, Codex, Cursor, or all supported harnesses. This updates only user-level config under the home directory; it never writes project files."),
			mcp.WithString(
				"harness",
				mcp.Required(),
				mcp.Enum("claude", "codex", "cursor", "all"),
				mcp.Description("Agent harness to configure"),
			),
		),
		handleHooksInstall,
	)
}

func handleHooksInstall(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("harness")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name = strings.ToLower(strings.TrimSpace(name))

	var changed bool
	if name == "all" {
		changed, err = hooks.InstallAll()
	} else {
		var harness hooks.Harness
		harness, err = hooks.ParseHarness(name)
		if err == nil {
			changed, err = hooks.Install(harness)
			name = string(harness)
		}
	}
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !changed {
		return mcp.NewToolResultText("global Loom hooks already installed, nothing to do"), nil
	}

	message := fmt.Sprintf("installed global Loom hooks for %s", name)
	if name == string(hooks.HarnessCodex) || name == "all" {
		message += " — in Codex, run /hooks to review and trust them"
	}
	return mcp.NewToolResultText(message), nil
}

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/nudoxorg/loom/internal/hooks"
	"github.com/nudoxorg/loom/internal/project"
)

func registerHooksTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("loom_hooks_install",
			mcp.WithDescription("Install project-scoped Claude Code hooks (SessionStart, PreToolUse) that remind any agent working in this project to use Loom's MCP tools"),
		),
		handleHooksInstall,
	)
}

func handleHooksInstall(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := project.EnsureHome(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	proj, err := project.Resolve(".")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	changed, err := hooks.Install(proj.Root)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !changed {
		return mcp.NewToolResultText("Loom hooks already installed, nothing to do"), nil
	}

	return mcp.NewToolResultText("installed Loom hooks in .claude/settings.json"), nil
}

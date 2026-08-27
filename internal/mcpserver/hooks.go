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

	s.AddTool(
		mcp.NewTool("loom_hooks_install_global",
			mcp.WithDescription("Install a single global Claude Code SessionStart hook (in the user's own ~/.claude/settings.json, not any project's) that nudges an agent to consider Loom whenever it's working inside a git repository"),
		),
		handleHooksInstallGlobal,
	)

	s.AddTool(
		mcp.NewTool("loom_hooks_install_codex",
			mcp.WithDescription("Install project-scoped Codex CLI hooks (SessionStart, PreToolUse) that remind any agent working in this project to use Loom's MCP tools"),
		),
		handleHooksInstallCodex,
	)

	s.AddTool(
		mcp.NewTool("loom_hooks_install_codex_global",
			mcp.WithDescription("Install a single global Codex CLI SessionStart hook (in the user's own ~/.codex/hooks.json, not any project's) that nudges an agent to consider Loom whenever it's working inside a git repository"),
		),
		handleHooksInstallCodexGlobal,
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

func handleHooksInstallGlobal(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	changed, err := hooks.InstallGlobal()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !changed {
		return mcp.NewToolResultText("global Loom hook already installed, nothing to do"), nil
	}

	return mcp.NewToolResultText("installed global Loom hook in ~/.claude/settings.json"), nil
}

func handleHooksInstallCodex(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := project.EnsureHome(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	proj, err := project.Resolve(".")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	changed, err := hooks.InstallCodex(proj.Root)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !changed {
		return mcp.NewToolResultText("Loom hooks already installed, nothing to do"), nil
	}

	return mcp.NewToolResultText("installed Loom hooks in .codex/hooks.json — open the Codex CLI and run /hooks to review and trust them before they'll fire"), nil
}

func handleHooksInstallCodexGlobal(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	changed, err := hooks.InstallCodexGlobal()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !changed {
		return mcp.NewToolResultText("global Loom hook already installed, nothing to do"), nil
	}

	return mcp.NewToolResultText("installed global Loom hook in ~/.codex/hooks.json — open the Codex CLI and run /hooks to review and trust it before it'll fire"), nil
}

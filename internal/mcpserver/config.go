package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/nudoxorg/loom/internal/config"
	"github.com/nudoxorg/loom/internal/project"
)

func registerConfigTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("loom_config_get",
			mcp.WithDescription("Get a single Loom config value"),
			mcp.WithString("key", mcp.Required(), mcp.Description("config key, e.g. default_agent or default_limit")),
		),
		handleConfigGet,
	)

	s.AddTool(
		mcp.NewTool("loom_config_list",
			mcp.WithDescription("List all Loom config values"),
		),
		handleConfigList,
	)
}

func handleConfigGet(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := project.EnsureHome(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	settings, err := config.Load()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	value, err := config.Get(settings, key)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(value), nil
}

func handleConfigList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := project.EnsureHome(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	settings, err := config.Load()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var b strings.Builder
	for _, key := range config.Keys() {
		value, err := config.Get(settings, key)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fmt.Fprintf(&b, "%s = %s\n", key, value)
	}

	return mcp.NewToolResultText(strings.TrimRight(b.String(), "\n")), nil
}

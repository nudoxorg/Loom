package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/nudoxorg/loom/internal/config"
	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
)

func registerEventTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool(
			"loom_log",
			mcp.WithDescription("Log an event to the current project's Loom history"),
			mcp.WithString("cwd", mcp.Required(), mcp.Description(cwdDescription)),
			mcp.WithString("message", mcp.Required(), mcp.Description("what happened")),
		),
		handleLog,
	)

	s.AddTool(
		mcp.NewTool(
			"loom_show",
			mcp.WithDescription("Show recent events for the current project"),
			mcp.WithString("cwd", mcp.Required(), mcp.Description(cwdDescription)),
			mcp.WithNumber("limit", mcp.Description("max events to return (default from config)")),
		),
		handleShow,
	)
}

func handleLog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cwd, err := req.RequireString("cwd")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	message, err := req.RequireString("message")
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

	proj, err := project.Resolve(cwd)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := storage.Open(proj.DBPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer storage.Close(db)

	event := storage.Event{
		Agent:     resolveAgent(ctx, settings),
		Kind:      "log",
		Message:   message,
		Timestamp: time.Now(),
	}

	if err := storage.InsertEvent(db, event); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("logged"), nil
}

func handleShow(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cwd, err := req.RequireString("cwd")
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

	proj, err := project.Resolve(cwd)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := storage.Open(proj.DBPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer storage.Close(db)

	limit := req.GetInt("limit", settings.DefaultLimit)
	if limit <= 0 {
		limit = settings.DefaultLimit
	}

	events, err := storage.ListEvents(db, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(formatEvents(events)), nil
}

func formatEvents(events []storage.Event) string {
	if len(events) == 0 {
		return "no events"
	}

	var b strings.Builder
	for _, e := range events {
		fmt.Fprintf(&b, "%s [%s] %s: %s\n", e.Timestamp.Format("2006-01-02 15:04"), e.Kind, e.Agent, e.Message)
		if e.Path != "" {
			fmt.Fprintf(&b, "\tpath: %s\n", e.Path)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

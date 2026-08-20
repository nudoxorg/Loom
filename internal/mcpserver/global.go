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

func registerGlobalTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("loom_global_log",
			mcp.WithDescription("Log an event to Loom's global (cross-project) context"),
			mcp.WithString("message", mcp.Required(), mcp.Description("what happened")),
		),
		handleGlobalLog,
	)

	s.AddTool(
		mcp.NewTool("loom_global_show",
			mcp.WithDescription("Show recent events from Loom's global (cross-project) context"),
			mcp.WithNumber("limit", mcp.Description("max events to return (default from config)")),
		),
		handleGlobalShow,
	)

	s.AddTool(
		mcp.NewTool("loom_global_all",
			mcp.WithDescription("Show active claims and recent events across every project Loom has touched, plus the global log"),
			mcp.WithNumber("limit", mcp.Description("max events to return (default from config)")),
		),
		handleGlobalAll,
	)
}

func handleGlobalLog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	globalDBPath, err := project.GlobalDBPath()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := storage.Open(globalDBPath)
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

func handleGlobalShow(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := project.EnsureHome(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	settings, err := config.Load()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	globalDBPath, err := project.GlobalDBPath()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := storage.Open(globalDBPath)
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

func handleGlobalAll(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := project.EnsureHome(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	settings, err := config.Load()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	limit := req.GetInt("limit", settings.DefaultLimit)
	if limit <= 0 {
		limit = settings.DefaultLimit
	}

	claims, events, err := project.AggregateAll(limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var b strings.Builder
	fmt.Fprintln(&b, "ACTIVE CLAIMS")
	fmt.Fprintln(&b, formatLabeledClaims(claims))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "RECENT EVENTS")
	fmt.Fprint(&b, formatLabeledEvents(events))

	return mcp.NewToolResultText(b.String()), nil
}

func formatLabeledEvents(events []project.LabeledEvent) string {
	if len(events) == 0 {
		return "no events"
	}

	var b strings.Builder
	for _, le := range events {
		e := le.Event
		fmt.Fprintf(&b, "%s  %s [%s] %s: %s\n", le.Project, e.Timestamp.Format("2006-01-02 15:04"), e.Kind, e.Agent, e.Message)
		if e.Path != "" {
			fmt.Fprintf(&b, "\tpath: %s\n", e.Path)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func formatLabeledClaims(claims []project.LabeledClaim) string {
	if len(claims) == 0 {
		return "no active claims"
	}

	var b strings.Builder
	for _, lc := range claims {
		c := lc.Claim
		fmt.Fprintf(&b, "%s  %s  %s: %s\n", lc.Project, c.ClaimedAt.Format("2006-01-02 15:04"), c.Agent, c.Path)
	}

	return strings.TrimRight(b.String(), "\n")
}

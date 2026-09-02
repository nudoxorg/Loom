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

func registerClaimTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool(
			"loom_claim",
			mcp.WithDescription("Claim a path in the current project as in progress"),
			mcp.WithString("cwd", mcp.Required(), mcp.Description(cwdDescription)),
			mcp.WithString("path", mcp.Required(), mcp.Description("the file or directory path being worked on")),
		),
		handleClaim,
	)

	s.AddTool(
		mcp.NewTool(
			"loom_release",
			mcp.WithDescription("Release a previously claimed path"),
			mcp.WithString("cwd", mcp.Required(), mcp.Description(cwdDescription)),
			mcp.WithString("path", mcp.Required(), mcp.Description("the path to release")),
		),
		handleRelease,
	)

	s.AddTool(
		mcp.NewTool(
			"loom_status",
			mcp.WithDescription("Show currently active claims for the current project"),
			mcp.WithString("cwd", mcp.Required(), mcp.Description(cwdDescription)),
		),
		handleStatus,
	)
}

func handleClaim(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cwd, err := req.RequireString("cwd")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	path, err := req.RequireString("path")
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

	path, err = project.NormalizePath(proj.Root, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := storage.Open(proj.DBPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer storage.Close(db)

	agent := resolveAgent(ctx, settings)

	claim := storage.Claim{
		Path:      path,
		Agent:     agent,
		ClaimedAt: time.Now(),
	}

	created, err := storage.InsertClaim(db, claim)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !created {
		return mcp.NewToolResultText("already claimed " + path), nil
	}

	event := storage.Event{
		Agent:     agent,
		Kind:      "claim",
		Message:   "claimed " + path,
		Path:      path,
		Timestamp: time.Now(),
	}

	if err := storage.InsertEvent(db, event); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("claimed " + path), nil
}

func handleRelease(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cwd, err := req.RequireString("cwd")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	path, err := req.RequireString("path")
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

	path, err = project.NormalizePath(proj.Root, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	db, err := storage.Open(proj.DBPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer storage.Close(db)

	agent := resolveAgent(ctx, settings)

	if err := storage.ReleaseClaim(db, path, agent); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	event := storage.Event{
		Agent:     agent,
		Kind:      "release",
		Message:   "released " + path,
		Path:      path,
		Timestamp: time.Now(),
	}

	if err := storage.InsertEvent(db, event); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("released " + path), nil
}

func handleStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cwd, err := req.RequireString("cwd")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := project.EnsureHome(); err != nil {
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

	claims, err := storage.ListActiveClaims(db)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if len(claims) == 0 {
		return mcp.NewToolResultText("no active claims"), nil
	}

	var b strings.Builder
	for _, c := range claims {
		fmt.Fprintf(&b, "%s  %s: %s\n", c.ClaimedAt.Format("2006-01-02 15:04"), c.Agent, c.Path)
	}

	return mcp.NewToolResultText(strings.TrimRight(b.String(), "\n")), nil
}

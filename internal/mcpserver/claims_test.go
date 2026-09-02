package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/nudoxorg/loom/internal/project"
)

func withMCPTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if err := project.EnsureHome(); err != nil {
		t.Fatalf("EnsureHome() error = %v", err)
	}
}

func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatalf("result has no content")
	}
	text, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("result content is %T, want mcp.TextContent", res.Content[0])
	}
	return text.Text
}

func callWith(args map[string]any) mcp.CallToolRequest {
	var req mcp.CallToolRequest
	req.Params.Arguments = args
	return req
}

func TestHandleClaimRequiresCwd(t *testing.T) {
	withMCPTempHome(t)

	res, err := handleClaim(context.Background(), callWith(map[string]any{"path": "foo.go"}))
	if err != nil {
		t.Fatalf("handleClaim() error = %v", err)
	}
	if !res.IsError {
		t.Fatalf("handleClaim() with no cwd should error, got %q", resultText(t, res))
	}
}

func TestHandleClaimResolvesAgainstGivenCwdNotProcessCwd(t *testing.T) {
	withMCPTempHome(t)

	repoA := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoA, ".git"), 0o700); err != nil {
		t.Fatalf("Mkdir(.git) error = %v", err)
	}

	// The test process's real cwd is irrelevant to resolution now; only the
	// cwd argument matters. Sanity-check they differ.
	if procCwd, _ := os.Getwd(); procCwd == repoA {
		t.Fatalf("test setup invalid: repoA equals process cwd")
	}

	claimRes, err := handleClaim(context.Background(), callWith(map[string]any{
		"cwd":  repoA,
		"path": "foo.go",
	}))
	if err != nil {
		t.Fatalf("handleClaim() error = %v", err)
	}
	if claimRes.IsError {
		t.Fatalf("handleClaim() unexpected error: %q", resultText(t, claimRes))
	}

	statusRes, err := handleStatus(context.Background(), callWith(map[string]any{"cwd": repoA}))
	if err != nil {
		t.Fatalf("handleStatus() error = %v", err)
	}
	if got := resultText(t, statusRes); got == "no active claims" {
		t.Fatalf("handleStatus(repoA) = %q, want the claim made against repoA", got)
	}

	// A different, unrelated directory must not see repoA's claim.
	repoB := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoB, ".git"), 0o700); err != nil {
		t.Fatalf("Mkdir(.git) error = %v", err)
	}
	otherStatus, err := handleStatus(context.Background(), callWith(map[string]any{"cwd": repoB}))
	if err != nil {
		t.Fatalf("handleStatus(repoB) error = %v", err)
	}
	if got := resultText(t, otherStatus); got != "no active claims" {
		t.Fatalf("handleStatus(repoB) = %q, want no active claims (separate project)", got)
	}
}

func TestHandleClaimAdHocProjectForNonGitCwd(t *testing.T) {
	withMCPTempHome(t)

	scratch := t.TempDir()

	claimRes, err := handleClaim(context.Background(), callWith(map[string]any{
		"cwd":  scratch,
		"path": "notes.md",
	}))
	if err != nil {
		t.Fatalf("handleClaim() error = %v", err)
	}
	if claimRes.IsError {
		t.Fatalf("handleClaim() in non-git dir should succeed via ad-hoc project, got error: %q", resultText(t, claimRes))
	}
}

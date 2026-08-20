// Package mcpserver exposes Loom's operations as MCP tools over stdio,
// mirroring internal/cli's commands one-for-one.
package mcpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/mark3labs/mcp-go/server"

	"github.com/nudoxorg/loom/internal/config"
)

const version = "0.1.0"

// instanceID distinguishes this server process from any other Loom MCP
// server running concurrently (e.g. a second Claude Code window). Each
// `loom mcp` invocation is its own stdio process serving exactly one
// client, so a random ID generated once at startup and folded into every
// agent identity this process reports is stable for the process's whole
// life but unique across processes.
var instanceID = generateInstanceID()

func generateInstanceID() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// New builds an MCP server exposing every Loom operation as a tool.
func New() *server.MCPServer {
	s := server.NewMCPServer("Loom", version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithInstructions(instructions),
	)

	registerEventTools(s)
	registerClaimTools(s)
	registerGlobalTools(s)
	registerConfigTools(s)

	return s
}

// resolveAgent identifies the calling agent: the connecting MCP client's
// self-reported name (e.g. "claude-code", "cursor") if available, otherwise
// the configured default_agent, suffixed with this process's instanceID.
//
// The suffix matters because every MCP client we've seen reports the same
// name for every session — two concurrent Claude Code windows both say
// "claude-code". Without a per-process suffix they'd be indistinguishable
// in the claims table, and worse, ReleaseClaim matches on (path, agent),
// so one window could silently release the other's claim on a shared
// path. The suffix keeps them as separate agents while events/claims
// within one window still read as one continuous identity.
func resolveAgent(ctx context.Context, settings config.Settings) string {
	name := settings.DefaultAgent

	session := server.ClientSessionFromContext(ctx)
	if session != nil {
		if withInfo, ok := session.(server.SessionWithClientInfo); ok {
			if clientName := withInfo.GetClientInfo().Name; clientName != "" {
				name = clientName
			}
		}
	}

	return name + ":" + instanceID
}

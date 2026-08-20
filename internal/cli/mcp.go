package cli

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/nudoxorg/loom/internal/mcpserver"
	"github.com/nudoxorg/loom/internal/project"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the Loom MCP server over stdio",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		return server.ServeStdio(mcpserver.New())
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

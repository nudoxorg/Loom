package cli

import (
	"github.com/nudoxorg/Loom/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "loom",
	Short:   "shared context for coding agents",
	Version: version.Version,
}

func Execute() error {
	return rootCmd.Execute()
}

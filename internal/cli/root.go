package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "loom",
	Short: "shared context for coding agents",
}

func Execute() error {
	return rootCmd.Execute()
}

package cli

import "github.com/spf13/cobra"

var globalCmd = &cobra.Command{
	Use:   "global",
	Short: "Interact with global context",
}

var globalLogCmd = &cobra.Command{
	Use:   "log <message>",
	Short: "Log an event to global context",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var globalShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show recent global events",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	globalCmd.AddCommand(globalLogCmd)
	globalCmd.AddCommand(globalShowCmd)
	rootCmd.AddCommand(globalCmd)
}

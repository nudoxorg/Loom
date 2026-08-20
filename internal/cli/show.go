package cli

import (
	"fmt"

	"github.com/nudoxorg/loom/internal/config"
	"github.com/nudoxorg/loom/internal/output"
	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
	"github.com/spf13/cobra"
)

var showLimit int

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show recent events for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		settings, err := config.Load()
		if err != nil {
			return err
		}

		proj, err := project.Resolve(".")
		if err != nil {
			return err
		}

		db, err := storage.Open(proj.DBPath)
		if err != nil {
			return err
		}
		defer storage.Close(db)

		limit := showLimit
		if limit <= 0 {
			limit = settings.DefaultLimit
		}

		events, err := storage.ListEvents(db, limit)
		if err != nil {
			return err
		}

		fmt.Println(output.Events(events))

		return nil
	},
}

func init() {
	showCmd.Flags().IntVarP(&showLimit, "limit", "l", 0, "number of events to show (default from config)")
	rootCmd.AddCommand(showCmd)
}

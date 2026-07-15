package cli

import (
	"time"

	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log <message>",
	Short: "Log an event to the current project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
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

		event := storage.Event{
			Agent:     "manual",
			Kind:      "log",
			Message:   args[0],
			Timestamp: time.Now(),
		}

		// idk which way to call this?
		// storage.InsertEvent(db, event)
		if err := storage.InsertEvent(db, event); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}

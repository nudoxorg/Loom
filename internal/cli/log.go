package cli

import (
	"fmt"
	"time"

	"github.com/nudoxorg/loom/internal/config"
	"github.com/nudoxorg/loom/internal/output"
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

		event := storage.Event{
			Agent:     settings.DefaultAgent,
			Kind:      "log",
			Message:   args[0],
			Timestamp: time.Now(),
		}

		if err := storage.InsertEvent(db, event); err != nil {
			return err
		}

		fmt.Println(output.Success("logged"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}

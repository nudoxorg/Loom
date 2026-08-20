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

var globalCmd = &cobra.Command{
	Use:   "global",
	Short: "Interact with global context",
}

var globalLogCmd = &cobra.Command{
	Use:   "log <message>",
	Short: "Log an event to global context",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		settings, err := config.Load()
		if err != nil {
			return err
		}

		globalDBPath, err := project.GlobalDBPath()
		if err != nil {
			return err
		}

		db, err := storage.Open(globalDBPath)
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

var globalShowLimit int

var globalShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show recent global events",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		settings, err := config.Load()
		if err != nil {
			return err
		}

		globalDBPath, err := project.GlobalDBPath()
		if err != nil {
			return err
		}

		db, err := storage.Open(globalDBPath)
		if err != nil {
			return err
		}
		defer storage.Close(db)

		limit := globalShowLimit
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
	globalShowCmd.Flags().IntVarP(&globalShowLimit, "limit", "l", 0, "number of events to show (default from config)")
	globalCmd.AddCommand(globalLogCmd)
	globalCmd.AddCommand(globalShowCmd)
	rootCmd.AddCommand(globalCmd)
}

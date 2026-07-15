package cli

import (
	"fmt"
	"time"

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
			Agent:     "manual",
			Kind:      "log",
			Message:   args[0],
			Timestamp: time.Now(),
		}

		if err := storage.InsertEvent(db, event); err != nil {
			return err
		}

		return nil
	},
}

var globalShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show recent global events",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
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

		// need to add parameter for limit instead of defaulting at 20
		events, err := storage.ListEvents(db, 20)
		if err != nil {
			return err
		}

		for _, e := range events {
			fmt.Printf("%s [%s] %s: %s\n", e.Timestamp.Format("2006-01-02 15:04"), e.Kind, e.Agent, e.Message)
			if e.Path != "" {
				fmt.Printf("	path: %s\n", e.Path)
			}
		}

		return nil
	},
}

func init() {
	globalCmd.AddCommand(globalLogCmd)
	globalCmd.AddCommand(globalShowCmd)
	rootCmd.AddCommand(globalCmd)
}

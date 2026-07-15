package cli

import (
	"fmt"

	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show recent events for the current project",
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
	rootCmd.AddCommand(showCmd)
}

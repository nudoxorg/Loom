package cli

import (
	"time"

	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
	"github.com/spf13/cobra"
)

var claimCmd = &cobra.Command{
	Use:   "claim <path>",
	Short: "Claim a path as in progress",
	Args:  cobra.ExactArgs(1),
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

		claim := storage.Claim{
			Path:      args[0],
			Agent:     "manual",
			ClaimedAt: time.Now(),
		}

		if err := storage.InsertClaim(db, claim); err != nil {
			return err
		}

		event := storage.Event{
			Agent:     "manual",
			Kind:      "claim",
			Message:   "claimed " + args[0],
			Path:      args[0],
			Timestamp: time.Now(),
		}

		if err := storage.InsertEvent(db, event); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(claimCmd)
}

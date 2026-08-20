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

var claimCmd = &cobra.Command{
	Use:   "claim <path>",
	Short: "Claim a path as in progress",
	Args:  cobra.ExactArgs(1),
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

		claim := storage.Claim{
			Path:      args[0],
			Agent:     settings.DefaultAgent,
			ClaimedAt: time.Now(),
		}

		created, err := storage.InsertClaim(db, claim)
		if err != nil {
			return err
		}

		if !created {
			fmt.Println(output.Success("already claimed " + args[0]))
			return nil
		}

		event := storage.Event{
			Agent:     settings.DefaultAgent,
			Kind:      "claim",
			Message:   "claimed " + args[0],
			Path:      args[0],
			Timestamp: time.Now(),
		}

		if err := storage.InsertEvent(db, event); err != nil {
			return err
		}

		fmt.Println(output.Success("claimed " + args[0]))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(claimCmd)
}

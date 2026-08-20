package cli

import (
	"fmt"

	"github.com/nudoxorg/loom/internal/output"
	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show currently active claims",
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

		claims, err := storage.ListActiveClaims(db)
		if err != nil {
			return err
		}

		fmt.Println(output.Claims(claims))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

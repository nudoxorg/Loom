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

var releaseCmd = &cobra.Command{
	Use:   "release <path>",
	Short: "Release a claimed path",
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

		path, err := project.NormalizePath(proj.Root, args[0])
		if err != nil {
			return err
		}

		db, err := storage.Open(proj.DBPath)
		if err != nil {
			return err
		}
		defer storage.Close(db)

		if err := storage.ReleaseClaim(db, path, settings.DefaultAgent); err != nil {
			return err
		}

		event := storage.Event{
			Agent:     settings.DefaultAgent,
			Kind:      "release",
			Message:   "released " + path,
			Path:      path,
			Timestamp: time.Now(),
		}

		if err := storage.InsertEvent(db, event); err != nil {
			return err
		}

		fmt.Println(output.Success("released " + path))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}

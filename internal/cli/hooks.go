package cli

import (
	"fmt"

	"github.com/nudoxorg/loom/internal/hooks"
	"github.com/nudoxorg/loom/internal/output"
	"github.com/nudoxorg/loom/internal/project"
	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage Claude Code hooks for this project",
}

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install project-scoped Claude Code hooks that remind agents to use Loom",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		proj, err := project.Resolve(".")
		if err != nil {
			return err
		}

		changed, err := hooks.Install(proj.Root)
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("Loom hooks already installed, nothing to do"))
			return nil
		}

		fmt.Println(output.Success("installed Loom hooks in .claude/settings.json"))

		return nil
	},
}

func init() {
	hooksCmd.AddCommand(hooksInstallCmd)
	rootCmd.AddCommand(hooksCmd)
}

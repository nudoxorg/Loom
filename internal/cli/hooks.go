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
	Short: "Manage coding-agent hooks (Claude Code, Codex CLI, Cursor, Antigravity) that remind agents to use Loom",
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

var hooksInstallGlobalCmd = &cobra.Command{
	Use:   "install-global",
	Short: "Install a global Claude Code hook that nudges agents to use Loom in any git repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		changed, err := hooks.InstallGlobal()
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("global Loom hook already installed, nothing to do"))
			return nil
		}

		fmt.Println(output.Success("installed global Loom hook in ~/.claude/settings.json"))

		return nil
	},
}

var hooksInstallCodexCmd = &cobra.Command{
	Use:   "install-codex",
	Short: "Install project-scoped Codex CLI hooks that remind agents to use Loom",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		proj, err := project.Resolve(".")
		if err != nil {
			return err
		}

		changed, err := hooks.InstallCodex(proj.Root)
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("Loom hooks already installed, nothing to do"))
			return nil
		}

		fmt.Println(output.Success("installed Loom hooks in .codex/hooks.json — open the Codex CLI and run /hooks to review and trust them before they'll fire"))

		return nil
	},
}

var hooksInstallCodexGlobalCmd = &cobra.Command{
	Use:   "install-codex-global",
	Short: "Install a global Codex CLI hook that nudges agents to use Loom in any git repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		changed, err := hooks.InstallCodexGlobal()
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("global Loom hook already installed, nothing to do"))
			return nil
		}

		fmt.Println(output.Success("installed global Loom hook in ~/.codex/hooks.json — open the Codex CLI and run /hooks to review and trust it before it'll fire"))

		return nil
	},
}

var hooksInstallCursorCmd = &cobra.Command{
	Use:   "install-cursor",
	Short: "Install project-scoped Cursor hooks that remind agents to use Loom",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		proj, err := project.Resolve(".")
		if err != nil {
			return err
		}

		changed, err := hooks.InstallCursor(proj.Root)
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("Loom hooks already installed, nothing to do"))
			return nil
		}

		fmt.Println(output.Success("installed Loom hooks in .cursor/hooks.json"))

		return nil
	},
}

var hooksInstallCursorGlobalCmd = &cobra.Command{
	Use:   "install-cursor-global",
	Short: "Install a global Cursor hook that nudges agents to use Loom in any git repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		changed, err := hooks.InstallCursorGlobal()
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("global Loom hook already installed, nothing to do"))
			return nil
		}

		fmt.Println(output.Success("installed global Loom hook in ~/.cursor/hooks.json"))

		return nil
	},
}

var hooksInstallAntigravityCmd = &cobra.Command{
	Use:   "install-antigravity",
	Short: "Install project-scoped Antigravity hooks that remind agents to use Loom",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.EnsureHome(); err != nil {
			return err
		}

		proj, err := project.Resolve(".")
		if err != nil {
			return err
		}

		changed, err := hooks.InstallAntigravity(proj.Root)
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("Loom hooks already installed, nothing to do"))
			return nil
		}

		fmt.Println(output.Success("installed Loom hooks in .agents/hooks.json"))

		return nil
	},
}

func init() {
	hooksCmd.AddCommand(hooksInstallCmd)
	hooksCmd.AddCommand(hooksInstallGlobalCmd)
	hooksCmd.AddCommand(hooksInstallCodexCmd)
	hooksCmd.AddCommand(hooksInstallCodexGlobalCmd)
	hooksCmd.AddCommand(hooksInstallCursorCmd)
	hooksCmd.AddCommand(hooksInstallCursorGlobalCmd)
	hooksCmd.AddCommand(hooksInstallAntigravityCmd)
	rootCmd.AddCommand(hooksCmd)
}

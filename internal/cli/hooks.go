package cli

import (
	"fmt"
	"strings"

	"github.com/nudoxorg/loom/internal/hooks"
	"github.com/nudoxorg/loom/internal/output"
	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage global coding-agent hooks that remind agents to use Loom",
}

var hooksInstallCmd = &cobra.Command{
	Use:       "install <claude|codex|cursor|all>",
	Short:     "Install global Loom reminder hooks for an agent harness",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"claude", "codex", "cursor", "all"},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.ToLower(strings.TrimSpace(args[0]))

		var (
			changed bool
			err     error
		)
		if name == "all" {
			changed, err = hooks.InstallAll()
		} else {
			var harness hooks.Harness
			harness, err = hooks.ParseHarness(name)
			if err == nil {
				changed, err = hooks.Install(harness)
				name = string(harness)
			}
		}
		if err != nil {
			return err
		}

		if !changed {
			fmt.Println(output.Success("global Loom hooks already installed, nothing to do"))
			return nil
		}

		message := fmt.Sprintf("installed global Loom hooks for %s", name)
		if name == string(hooks.HarnessCodex) || name == "all" {
			message += " — in Codex, run /hooks to review and trust them"
		}
		fmt.Println(output.Success(message))
		return nil
	},
}

func init() {
	hooksCmd.AddCommand(hooksInstallCmd)
	rootCmd.AddCommand(hooksCmd)
}

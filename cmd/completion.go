package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell autocompletion scripts",
	Long: `Generate shell completion script for the specified shell.

To enable completion for the current shell session:

  source <(drp completion bash)

To make it permanent (bash):

  drp completion bash | sudo tee /etc/bash_completion.d/drp

To make it permanent (zsh):

  drp completion zsh > "${fpath[1]}/_drp"`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(c *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("completion: unsupported shell %q", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

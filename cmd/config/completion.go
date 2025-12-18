package config

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
)

const shellFish = "fish"

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate completion script for the specified shell.

To configure shell completion:

Bash:
  $ workspace config completion bash > /etc/bash_completion.d/workspace

Zsh:
  $ workspace config completion zsh > "${fpath[1]}/_workspace"

Fish:
  $ workspace config completion fish > ~/.config/fish/completions/workspace.fish

PowerShell:
  PS> workspace config completion powershell > workspace.ps1`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(_ *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.RootCmd.GenZshCompletion(os.Stdout)
		case shellFish:
			err = cmd.RootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.RootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(completionCmd)
}

package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/ui/setup"
	"github.com/jcleira/workspace/pkg/workspace"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive configuration setup",
	Long:  `Interactively configure workspace directories and settings.`,
	Run: func(_ *cobra.Command, _ []string) {
		runInteractiveSetup()
	},
}

func init() {
	ConfigCmd.AddCommand(setupCmd)
}

func runInteractiveSetup() {
	cfg := cmd.ConfigManager.GetConfig()

	result, err := setup.RunSetupWizardWithDefaults(
		cmd.ConfigManager,
		cfg.ReposDir,
		cfg.WorkspacesDir,
		cfg.ClaudeDir,
	)
	if err != nil {
		commands.PrintError(fmt.Sprintf("Setup wizard failed: %v", err))
		os.Exit(1)
	}

	if !result.Completed {
		return
	}

	cfg = cmd.ConfigManager.GetConfig()
	cmd.WorkspaceManager = workspace.NewManager(cfg.WorkspacesDir, cfg.ReposDir, cfg.ClaudeDir)

	fmt.Println()
	showConfig()
}

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/workspace"
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long:  `Set a configuration value. Available keys: workspaces-dir, repos-dir, claude-dir`,
	Example: `  workspace config set repos-dir ~/Projects/repos
  workspace config set workspaces-dir ~/dev/workspaces
  workspace config set claude-dir ~/shared/.claude`,
	Args: cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		setConfigValue(args[0], args[1])
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}

func setConfigValue(key, value string) {
	switch key {
	case "workspaces-dir":
		if err := cmd.ConfigManager.SetWorkspacesDir(value); err != nil {
			commands.PrintErrorf("Failed to set workspaces directory: %v", err)
			return
		}
		cfg := cmd.ConfigManager.GetConfig()
		cmd.WorkspaceManager = workspace.NewManager(cfg.WorkspacesDir, cfg.ReposDir, cfg.ClaudeDir)
		commands.PrintSuccessf("Workspaces directory set to: %s", cfg.WorkspacesDir)

	case "repos-dir":
		if err := cmd.ConfigManager.SetReposDir(value); err != nil {
			commands.PrintErrorf("Failed to set repos directory: %v", err)
			return
		}
		cfg := cmd.ConfigManager.GetConfig()
		cmd.WorkspaceManager = workspace.NewManager(cfg.WorkspacesDir, cfg.ReposDir, cfg.ClaudeDir)
		commands.PrintSuccessf("Repos directory set to: %s", cfg.ReposDir)

	case "claude-dir":
		if err := cmd.ConfigManager.SetClaudeDir(value); err != nil {
			commands.PrintErrorf("Failed to set claude directory: %v", err)
			return
		}
		cfg := cmd.ConfigManager.GetConfig()
		cmd.WorkspaceManager = workspace.NewManager(cfg.WorkspacesDir, cfg.ReposDir, cfg.ClaudeDir)
		commands.PrintSuccessf("Claude directory set to: %s", cfg.ClaudeDir)

	default:
		commands.PrintErrorf("Unknown configuration key: %s", key)
		fmt.Println("Available keys: workspaces-dir, repos-dir, claude-dir")
	}
}

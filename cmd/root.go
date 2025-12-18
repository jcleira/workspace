// Package cmd provides the CLI commands for the workspace tool.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/pkg/config"
	"github.com/jcleira/workspace/pkg/shell"
	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/ui/dashboard"
	"github.com/jcleira/workspace/pkg/workspace"
)

var (
	ConfigManager    *config.ConfigManager
	WorkspaceManager *workspace.Manager
	OutputPathOnly   bool
)

var RootCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Workspace manager for multiple project development",
	Long:  `A CLI tool to create and manage isolated workspaces with multiple git repositories.`,
	Run: func(_ *cobra.Command, _ []string) {
		runInteractiveWorkspaceSelector()
	},
}

// InitializeConfig loads configuration and creates the workspace manager.
func InitializeConfig() error {
	var err error
	ConfigManager, err = config.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	cfg := ConfigManager.GetConfig()
	WorkspaceManager = workspace.NewManager(cfg.WorkspacesDir, cfg.ReposDir, cfg.ClaudeDir)

	return nil
}

func init() {
	RootCmd.Flags().BoolVar(&OutputPathOnly, "output-path-only", false, "Output only the selected workspace path (for shell integration)")
}

func runInteractiveWorkspaceSelector() {
	workspaces, err := WorkspaceManager.GetWorkspaces()
	if err != nil {
		commands.PrintError(fmt.Sprintf("Failed to get workspaces: %v", err))
		return
	}

	if len(workspaces) == 0 {
		commands.PrintWarning("No workspaces found.")
		fmt.Println("Create one with: workspace create <name>")
		return
	}

	selectedPath, err := dashboard.RunDashboard(WorkspaceManager, ConfigManager)
	if err != nil {
		commands.PrintError(fmt.Sprintf("Dashboard error: %v", err))
		return
	}

	if selectedPath != "" {
		if OutputPathOnly {
			fmt.Println(selectedPath)
		} else {
			ws := workspace.Workspace{Path: selectedPath}
			shell.NavigateToWorkspace(ws)
		}
	} else if OutputPathOnly {
		fmt.Println("quit")
	}
}

// WorkspaceCompletionFunc provides shell completion for workspace names.
func WorkspaceCompletionFunc(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	workspaces, err := WorkspaceManager.GetWorkspaces()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, ws := range workspaces {
		name := ws.Name[10:]
		if toComplete == "" || (len(name) >= len(toComplete) && name[:len(toComplete)] == toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

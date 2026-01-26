// Package cmd provides the CLI commands for the workspace tool.
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/pkg/config"
	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/ui/dashboard"
	"github.com/jcleira/workspace/pkg/workspace"
)

var (
	ConfigManager    *config.ConfigManager
	WorkspaceManager *workspace.Manager
	OutputPathOnly   bool

	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion sets the version information for the CLI.
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	RootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

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
	RootCmd.SetVersionTemplate("workspace {{.Version}}\n")
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

	if OutputPathOnly {
		runOutputPathOnlyMode()
		return
	}

	if !isClaudeInstalled() {
		commands.PrintError("'claude' command not found")
		fmt.Println("Install: https://claude.ai/code")
		return
	}

	for {
		result, err := dashboard.RunDashboard(WorkspaceManager, ConfigManager)
		if err != nil {
			commands.PrintError(fmt.Sprintf("Dashboard error: %v", err))
			return
		}

		if result.Path == "" {
			return
		}

		if result.LaunchShell {
			launchShell(result.Path)
		} else {
			launchClaude(result.Path)
		}
	}
}

func runOutputPathOnlyMode() {
	result, err := dashboard.RunDashboard(WorkspaceManager, ConfigManager)
	if err != nil {
		commands.PrintError(fmt.Sprintf("Dashboard error: %v", err))
		return
	}

	if result.Path != "" {
		fmt.Println(result.Path)
	} else {
		fmt.Println("quit")
	}
}

func isClaudeInstalled() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func launchClaude(workspacePath string) {
	cmd := exec.Command("claude", "--continue")
	cmd.Dir = workspacePath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Run()
}

func launchShell(workspacePath string) {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/sh"
	}

	cmd := exec.Command(shellPath)
	cmd.Dir = workspacePath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Run()
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

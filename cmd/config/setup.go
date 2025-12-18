package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
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
	reader := bufio.NewReader(os.Stdin)
	homeDir := os.Getenv("HOME")

	fmt.Println()
	fmt.Println(commands.TitleStyle.Render("Workspace Configuration Setup"))
	fmt.Println()
	commands.PrintInfo("Configure the directories where workspace will store its data.")
	fmt.Println()

	fmt.Printf("Workspaces directory (where workspace projects will be created):\n")
	fmt.Printf("Default: %s\n", commands.InfoStyle.Render(filepath.Join(homeDir, "workspaces")))
	fmt.Print("Enter path: ")
	workspacesDir, err := reader.ReadString('\n')
	if err != nil || strings.TrimSpace(workspacesDir) == "" {
		workspacesDir = filepath.Join(homeDir, "workspaces")
	} else {
		workspacesDir = strings.TrimSpace(workspacesDir)
	}

	fmt.Println()
	fmt.Printf("Repositories directory (where git repositories will be cloned):\n")
	fmt.Printf("Default: %s\n", commands.InfoStyle.Render(filepath.Join(homeDir, "repos")))
	fmt.Print("Enter path: ")
	reposDir, err := reader.ReadString('\n')
	if err != nil || strings.TrimSpace(reposDir) == "" {
		reposDir = filepath.Join(homeDir, "repos")
	} else {
		reposDir = strings.TrimSpace(reposDir)
	}

	fmt.Println()
	fmt.Printf("Claude directory (shared .claude context directory):\n")
	fmt.Printf("Default: %s\n", commands.InfoStyle.Render(filepath.Join(homeDir, ".claude")))
	fmt.Print("Enter path: ")
	claudeDir, err := reader.ReadString('\n')
	if err != nil || strings.TrimSpace(claudeDir) == "" {
		claudeDir = filepath.Join(homeDir, ".claude")
	} else {
		claudeDir = strings.TrimSpace(claudeDir)
	}

	fmt.Println()
	commands.PrintInfo("Saving configuration...")

	if err := cmd.ConfigManager.SetWorkspacesDir(workspacesDir); err != nil {
		commands.PrintError(fmt.Sprintf("Failed to set workspaces directory: %v", err))
		os.Exit(1)
	}

	if err := cmd.ConfigManager.SetReposDir(reposDir); err != nil {
		commands.PrintError(fmt.Sprintf("Failed to set repos directory: %v", err))
		os.Exit(1)
	}

	if err := cmd.ConfigManager.SetClaudeDir(claudeDir); err != nil {
		commands.PrintError(fmt.Sprintf("Failed to set claude directory: %v", err))
		os.Exit(1)
	}

	cfg := cmd.ConfigManager.GetConfig()
	cmd.WorkspaceManager = workspace.NewManager(cfg.WorkspacesDir, cfg.ReposDir, cfg.ClaudeDir)

	fmt.Println()
	commands.PrintSuccess("Configuration saved successfully!")
	fmt.Println()
	showConfig()
}

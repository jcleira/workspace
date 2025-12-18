// Package list provides the workspace list command.
package list

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/workspace"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `Display all existing workspaces and their contents.`,
	Run: func(_ *cobra.Command, _ []string) {
		ListWorkspaces()
	},
}

func init() {
	cmd.RootCmd.AddCommand(listCmd)
}

func ListWorkspaces() {
	commands.PrintInfo("Available workspaces:")

	workspaces, err := cmd.WorkspaceManager.GetWorkspaces()
	if err != nil {
		commands.PrintError(fmt.Sprintf("Failed to get workspaces: %v", err))
		return
	}

	if len(workspaces) == 0 {
		commands.PrintWarning("No workspaces found")
		return
	}

	for _, ws := range workspaces {
		workspaceName := ws.Name

		if workspaceName == workspace.DefaultWorkspaceName {
			fmt.Printf("  %s %s %s\n", commands.SuccessStyle.Render("•"), workspaceName, commands.InfoStyle.Render("(protected)"))
		} else {
			fmt.Printf("  %s %s\n", commands.SuccessStyle.Render("•"), workspaceName)
		}

		infoPath := filepath.Join(ws.Path, ".workspace-info")
		if file, err := os.Open(infoPath); err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				fmt.Printf("      %s\n", scanner.Text())
			}
			_ = file.Close()
		}

		repoCount := 0
		if entries, err := os.ReadDir(ws.Path); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() != ".claude" {
					entryPath := filepath.Join(ws.Path, entry.Name())
					if _, err := os.Stat(filepath.Join(entryPath, ".git")); err == nil {
						fmt.Printf("      %s %s (git repo)\n", commands.InfoStyle.Render("└─"), entry.Name())
						repoCount++
					}
				}
			}
		}

		if repoCount == 0 {
			fmt.Println("      No repositories cloned yet")
		}
		fmt.Println()
	}
}

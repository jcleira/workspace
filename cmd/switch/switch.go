// Package switchcmd provides the workspace switch command.
package switchcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	listpkg "github.com/jcleira/workspace/cmd/list"
	"github.com/jcleira/workspace/pkg/ui/commands"
)

var switchCmd = &cobra.Command{
	Use:     "switch <name>",
	Aliases: []string{"sw"},
	Short:   "Switch to a workspace",
	Long:    `Show the path to switch to a workspace directory.`,
	Example: `  workspace switch myfeature
  workspace sw default`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: cmd.WorkspaceCompletionFunc,
	Run: func(_ *cobra.Command, args []string) {
		name := args[0]
		switchWorkspace(name)
	},
}

func init() {
	cmd.RootCmd.AddCommand(switchCmd)
}

func switchWorkspace(name string) {
	workspacePath, err := cmd.WorkspaceManager.GetWorkspacePath(name)
	if err != nil {
		commands.PrintErrorf("Workspace 'workspace-%s' not found", name)
		commands.PrintInfo("Run 'workspace list' to see available workspaces")
		commands.PrintInfof("Run 'workspace create %s' to create a new workspace", name)
		fmt.Println()
		fmt.Println("Available workspaces:")
		listpkg.ListWorkspaces()
		return
	}

	commands.PrintSuccessf("Switching to workspace: workspace-%s", name)
	fmt.Printf("cd %s\n", workspacePath)
}

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
	Use:               "switch <name>",
	Short:             "Switch to a workspace",
	Long:              `Show the path to switch to a workspace directory.`,
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
		commands.PrintError(err.Error())
		fmt.Println("Available workspaces:")
		listpkg.ListWorkspaces()
		return
	}

	commands.PrintSuccess(fmt.Sprintf("Switching to workspace: workspace-%s", name))
	fmt.Printf("cd %s\n", workspacePath)
}

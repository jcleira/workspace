// Package delete provides the workspace delete command.
package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/workspace"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"d", "rm"},
	Short:   "Delete a workspace",
	Long:    `Delete a workspace and all its contents, including associated git branches.`,
	Example: `  workspace delete myfeature
  workspace rm bugfix-123`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: cmd.WorkspaceCompletionFunc,
	Run: func(_ *cobra.Command, args []string) {
		name := args[0]
		deleteWorkspace(name)
	},
}

func init() {
	cmd.RootCmd.AddCommand(deleteCmd)
}

func deleteWorkspace(name string) {
	if name == "default" {
		commands.PrintError("The 'default' workspace is protected and cannot be deleted")
		commands.PrintInfo("The default workspace is reserved for persistent projects")
		return
	}

	svc := workspace.NewService(cmd.WorkspaceManager)

	if _, err := svc.GetPath(name); err != nil {
		commands.PrintErrorf("Workspace 'workspace-%s' not found", name)
		return
	}

	commands.PrintWarning("This will delete the workspace and all its contents, including branches!")

	if !commands.PromptYesNo(fmt.Sprintf("Are you sure you want to delete 'workspace-%s'? (y/n): ", name)) {
		commands.PrintInfo("Deletion canceled")
		return
	}

	input := workspace.DeleteInput{
		Name:           name,
		DeleteBranches: true,
	}

	output, err := svc.Delete(input)
	if err != nil {
		commands.PrintError(err.Error())
		return
	}

	displayDeleteResults(output)

	defaultPath, err := svc.GetPath("default")
	if err != nil {
		commands.PrintErrorf("Could not navigate to default workspace: %v", err)
		return
	}

	commands.PrintSuccessf("Workspace 'workspace-%s' deleted successfully", name)
	commands.PrintInfo("Switching to default workspace")
	fmt.Printf("cd %s\n", defaultPath)
}

func displayDeleteResults(output workspace.DeleteOutput) {
	for _, b := range output.DeletedBranches {
		commands.PrintSuccessf("Deleted branch '%s' in %s", b.BranchName, b.RepoName)
	}

	for _, b := range output.SkippedBranches {
		if b.Error != nil {
			commands.PrintWarningf("Failed to delete branch '%s' in %s: %v", b.BranchName, b.RepoName, b.Error)
		}
	}
}

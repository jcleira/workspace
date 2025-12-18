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
	Use:               "delete <name>",
	Short:             "Delete a workspace",
	Long:              `Delete a workspace and all its contents, including associated git branches.`,
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
		commands.PrintError(fmt.Sprintf("Workspace 'workspace-%s' not found", name))
		return
	}

	commands.PrintWarning("This will delete the workspace and all its contents!")
	commands.PrintInfo("Note: Associated branches will be deleted (branches with unpushed commits will be skipped)")

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
		commands.PrintError("Could not navigate to default workspace: " + err.Error())
		return
	}

	commands.PrintSuccess(fmt.Sprintf("Workspace 'workspace-%s' deleted successfully", name))
	commands.PrintInfo("Switching to default workspace")
	fmt.Printf("cd %s\n", defaultPath)
}

func displayDeleteResults(output *workspace.DeleteOutput) {
	for _, b := range output.DeletedBranches {
		commands.PrintSuccess(fmt.Sprintf("Deleted branch '%s' in %s", b.BranchName, b.RepoName))
	}

	for _, b := range output.SkippedBranches {
		if b.UnpushedCount > 0 {
			commands.PrintWarning(fmt.Sprintf("Skipped branch '%s' in %s (%d unpushed commits)", b.BranchName, b.RepoName, b.UnpushedCount))
		} else if b.Error != nil {
			commands.PrintWarning(fmt.Sprintf("Failed to delete branch '%s' in %s: %v", b.BranchName, b.RepoName, b.Error))
		}
	}
}

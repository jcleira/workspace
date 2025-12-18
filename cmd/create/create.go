// Package create provides the workspace create command.
package create

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/shell"
	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/workspace"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new workspace",
	Long:  `Create a new workspace directory using all repositories in the repos directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		name := args[0]
		createWorkspace(name)
	},
}

func init() {
	cmd.RootCmd.AddCommand(createCmd)
}

func createWorkspace(name string) {
	svc := workspace.NewService(cmd.WorkspaceManager)

	if workspacePath, err := svc.GetPath(name); err == nil {
		commands.PrintWarning(fmt.Sprintf("Workspace 'workspace-%s' already exists", name))
		if !commands.PromptYesNo("Do you want to use the existing workspace? (y/n): ") {
			commands.PrintInfo("Creation canceled")
			return
		}
		commands.PrintSuccess(fmt.Sprintf("Using existing workspace at: %s", workspacePath))
		ws := workspace.Workspace{Name: name, Path: workspacePath}
		shell.NavigateToWorkspace(ws)
		return
	}

	input := workspace.CreateInput{
		Name: name,
	}

	commands.PrintInfo(fmt.Sprintf("Creating workspace '%s'...", name))

	output, err := svc.Create(input)
	if err != nil {
		commands.PrintError(fmt.Sprintf("Failed to create workspace: %v", err))
		os.Exit(1)
	}

	displayCreateResults(output, name)

	ws := workspace.Workspace{
		Name: name,
		Path: output.WorkspacePath,
	}

	shell.NavigateToWorkspace(ws)
}

func displayCreateResults(output *workspace.CreateOutput, name string) {
	if output.AlreadyExists {
		commands.PrintWarning(fmt.Sprintf("Workspace 'workspace-%s' already exists", name))
	} else {
		commands.PrintSuccess(fmt.Sprintf("Workspace created at: %s", output.WorkspacePath))
	}

	for _, sync := range output.SyncResults {
		if sync.Error != nil {
			commands.PrintWarning(fmt.Sprintf("Failed to sync %s: %v", sync.RepoName, sync.Error))
		} else if sync.Fetched && sync.Pulled {
			commands.PrintSuccess(fmt.Sprintf("Updated %s", sync.RepoName))
		}
	}

	for _, repo := range output.CreatedRepos {
		if output.WorkspaceType == workspace.WorkspaceTypeWorktree {
			if repo.WasExisting {
				commands.PrintInfo(fmt.Sprintf("Checked out existing branch '%s' for %s", repo.BranchName, repo.Name))
			} else {
				commands.PrintSuccess(fmt.Sprintf("Created worktree for %s (new branch: %s)", repo.Name, repo.BranchName))
			}
		} else {
			commands.PrintSuccess(fmt.Sprintf("Cloned %s", repo.Name))
		}
	}

	for _, repo := range output.FailedRepos {
		commands.PrintError(fmt.Sprintf("Failed to create %s: %v", repo.Name, repo.Error))
	}

	if len(output.CreatedRepos) > 0 {
		if output.WorkspaceType == workspace.WorkspaceTypeWorktree {
			commands.PrintSuccess(fmt.Sprintf("Created %d worktrees!", len(output.CreatedRepos)))
		} else {
			commands.PrintSuccess(fmt.Sprintf("Cloned %d repositories!", len(output.CreatedRepos)))
		}
	}

	if len(output.FailedRepos) > 0 {
		commands.PrintWarning(fmt.Sprintf("%d repositories failed", len(output.FailedRepos)))
	}

	commands.PrintSuccess(fmt.Sprintf("Workspace 'workspace-%s' is ready!", name))
}

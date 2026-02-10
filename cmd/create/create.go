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
	Use:     "create <name>",
	Aliases: []string{"c"},
	Short:   "Create a new workspace",
	Long:    `Create a new workspace directory using all repositories in the repos directory.`,
	Example: `  workspace create myfeature
  workspace create bugfix-123
  workspace c quick-test`,
	Args: cobra.ExactArgs(1),
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
		commands.PrintWarningf("Workspace 'workspace-%s' already exists", name)
		if !commands.PromptYesNo("Do you want to use the existing workspace? (y/n): ") {
			commands.PrintInfo("Creation canceled")
			return
		}
		commands.PrintSuccessf("Using existing workspace at: %s", workspacePath)
		ws := workspace.Workspace{Name: name, Path: workspacePath}
		shell.NavigateToWorkspace(ws)
		return
	}

	spinner := commands.NewSpinner(fmt.Sprintf("Creating workspace '%s'...", name))
	spinner.Start()

	input := workspace.CreateInput{
		Name: name,
		OnProgress: func(message string) {
			spinner.UpdateMessage(message)
		},
	}

	output, err := svc.Create(input)
	spinner.Stop()

	if err != nil {
		commands.PrintErrorf("Failed to create workspace: %v", err)
		if os.IsNotExist(err) || fmt.Sprintf("%v", err) == fmt.Sprintf("no repositories found in %s", cmd.WorkspaceManager.ReposDir) {
			commands.PrintInfo("Run 'workspace config setup' to configure your directories")
		}
		os.Exit(1)
	}

	displayCreateResults(output, name)

	ws := workspace.Workspace{
		Name: name,
		Path: output.WorkspacePath,
	}

	shell.NavigateToWorkspace(ws)
}

func displayCreateResults(output workspace.CreateOutput, name string) {
	if output.AlreadyExists {
		commands.PrintWarningf("Workspace 'workspace-%s' already exists", name)
	} else {
		commands.PrintSuccessf("Created workspace at %s", output.WorkspacePath)
	}

	for _, sync := range output.SyncResults {
		if sync.Error != nil {
			commands.PrintWarningf("Failed to sync %s: %v", sync.RepoName, sync.Error)
		} else if sync.Fetched && sync.Pulled {
			commands.PrintSuccessf("Synced %s", sync.RepoName)
		}
	}

	for _, repo := range output.CreatedRepos {
		if output.WorkspaceType == workspace.WorkspaceTypeWorktree {
			if repo.WasExisting {
				commands.PrintInfof("Checked out existing branch '%s' for %s", repo.BranchName, repo.Name)
			} else {
				commands.PrintSuccessf("Created worktree for %s (branch: %s)", repo.Name, repo.BranchName)
			}
		} else {
			commands.PrintSuccessf("Cloned %s", repo.Name)
		}
	}

	for _, repo := range output.FailedRepos {
		commands.PrintErrorf("Failed to create %s: %v", repo.Name, repo.Error)
	}

	if len(output.CreatedRepos) == 0 && len(output.FailedRepos) == 0 {
		commands.PrintWarning("No repositories found in repos directory")
		commands.PrintInfo("Workspace created but contains no repos")
		commands.PrintInfo("Add repositories to the repos directory and recreate")
	} else if len(output.CreatedRepos) > 0 {
		if output.WorkspaceType == workspace.WorkspaceTypeWorktree {
			commands.PrintSuccessf("Created %d worktree(s)", len(output.CreatedRepos))
		} else {
			commands.PrintSuccessf("Cloned %d repository/ies", len(output.CreatedRepos))
		}
	}

	if len(output.FailedRepos) > 0 {
		commands.PrintWarningf("%d repository/ies failed", len(output.FailedRepos))
	}

	commands.PrintSuccessf("Workspace 'workspace-%s' is ready", name)
}

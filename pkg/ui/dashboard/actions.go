// Package dashboard provides the interactive terminal dashboard UI.
package dashboard

import (
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcleira/workspace/pkg/git"
)

// executeAction executes a git action on all repos in a workspace.
func executeAction(action ActionType, workspacePath string, projects []string, specificRepo string) tea.Cmd {
	return func() tea.Msg {
		var lastErr error
		var output string

		repos := projects
		if specificRepo != "" {
			repos = []string{specificRepo}
		}

		for _, project := range repos {
			repoPath := filepath.Join(workspacePath, project)

			switch action {
			case ActionFetch:
				if err := git.FetchRemote(repoPath); err != nil {
					lastErr = err
				}
			case ActionPull:
				if err := git.PullDefaultBranch(repoPath); err != nil {
					lastErr = err
				}
			case ActionDelete, ActionCreate, ActionSwitch, ActionDiff:
			}
		}

		return ActionCompleteMsg{
			Action:    action,
			Workspace: workspacePath,
			Repo:      specificRepo,
			Success:   lastErr == nil,
			Error:     lastErr,
			Output:    output,
		}
	}
}

// FetchAction creates a fetch command for a single repo.
func FetchAction(repoPath, repoName string) tea.Cmd {
	return func() tea.Msg {
		err := git.FetchRemote(repoPath)
		return ActionCompleteMsg{
			Action:  ActionFetch,
			Repo:    repoName,
			Success: err == nil,
			Error:   err,
		}
	}
}

// PullAction creates a pull command for a single repo.
func PullAction(repoPath, repoName string) tea.Cmd {
	return func() tea.Msg {
		err := git.PullDefaultBranch(repoPath)
		return ActionCompleteMsg{
			Action:  ActionPull,
			Repo:    repoName,
			Success: err == nil,
			Error:   err,
		}
	}
}

// executeDiffAction creates a command to show git diff using difftastic if available.
func executeDiffAction(repoPath string, staged bool) tea.Cmd {
	var cmd *exec.Cmd

	_, err := exec.LookPath("difft")
	hasDifft := err == nil

	if hasDifft {
		if staged {
			cmd = exec.Command("git", "-c", "diff.external=difft", "diff", "--staged")
		} else {
			cmd = exec.Command("git", "-c", "diff.external=difft", "diff")
		}
	} else {
		if staged {
			cmd = exec.Command("git", "diff", "--staged")
		} else {
			cmd = exec.Command("git", "diff")
		}
	}

	cmd.Dir = repoPath

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return nil
	})
}

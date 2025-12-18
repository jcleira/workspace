package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorktreeInfo contains information about a git worktree.
type WorktreeInfo struct {
	Path   string
	Branch string
	Commit string
}

// IsWorktree checks if the given path is a git worktree (not a regular clone).
func IsWorktree(path string) (bool, error) {
	gitPath := filepath.Join(path, ".git")

	info, err := os.Stat(gitPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !info.IsDir(), nil
}

// CreateWorktree creates a new git worktree with a new branch at the specified path.
func CreateWorktree(mainRepoPath, worktreePath, branchName string) error {
	cmd := exec.Command("git", "-C", mainRepoPath, "worktree", "add", "-b", branchName, worktreePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if strings.Contains(errorMsg, "already registered worktree") || strings.Contains(errorMsg, "missing but already registered") {
			if pruneErr := PruneWorktrees(mainRepoPath); pruneErr == nil {
				cmd = exec.Command("git", "-C", mainRepoPath, "worktree", "add", "-b", branchName, worktreePath)
				output, err = cmd.CombinedOutput()
				if err == nil {
					return nil
				}
				errorMsg = strings.TrimSpace(string(output))
			}
		}
		if strings.Contains(errorMsg, "already exists") {
			return fmt.Errorf("branch '%s' already exists: %w", branchName, err)
		}
		return fmt.Errorf("failed to create worktree: %s: %w", errorMsg, err)
	}

	return nil
}

// CheckoutExistingBranch creates a worktree for an existing branch at the specified path.
func CheckoutExistingBranch(mainRepoPath, worktreePath, branchName string) error {
	cmd := exec.Command("git", "-C", mainRepoPath, "worktree", "add", worktreePath, branchName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if strings.Contains(errorMsg, "already registered worktree") || strings.Contains(errorMsg, "missing but already registered") {
			if pruneErr := PruneWorktrees(mainRepoPath); pruneErr == nil {
				cmd = exec.Command("git", "-C", mainRepoPath, "worktree", "add", worktreePath, branchName)
				output, err = cmd.CombinedOutput()
				if err == nil {
					return nil
				}
				errorMsg = strings.TrimSpace(string(output))
			}
		}
		if strings.Contains(errorMsg, "already checked out") {
			return fmt.Errorf("branch '%s' is already checked out: %w", branchName, err)
		}
		return fmt.Errorf("failed to checkout existing branch: %s: %w", errorMsg, err)
	}

	return nil
}

// RemoveWorktree removes a git worktree from the repository.
func RemoveWorktree(mainRepoPath, worktreePath string) error {
	cmd := exec.Command("git", "-C", mainRepoPath, "worktree", "remove", worktreePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to remove worktree: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// PruneWorktrees removes stale worktree entries from the repository.
func PruneWorktrees(mainRepoPath string) error {
	cmd := exec.Command("git", "-C", mainRepoPath, "worktree", "prune")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to prune worktrees: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// ListWorktrees returns a list of all worktrees associated with the repository.
func ListWorktrees(mainRepoPath string) ([]WorktreeInfo, error) {
	cmd := exec.Command("git", "-C", mainRepoPath, "worktree", "list", "--porcelain")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var worktrees []WorktreeInfo
	var current WorktreeInfo

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch ")
		case strings.HasPrefix(line, "HEAD "):
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// GetMainRepoFromWorktree returns the path to the main repository for a given worktree.
func GetMainRepoFromWorktree(worktreePath string) (string, error) {
	gitFilePath := filepath.Join(worktreePath, ".git")

	content, err := os.ReadFile(gitFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read .git file: %w", err)
	}

	line := strings.TrimSpace(string(content))
	if !strings.HasPrefix(line, "gitdir: ") {
		return "", fmt.Errorf("invalid .git file format")
	}

	gitdir := strings.TrimPrefix(line, "gitdir: ")
	gitdir = filepath.Clean(gitdir)

	worktreesDir := filepath.Dir(gitdir)
	mainGitDir := filepath.Dir(worktreesDir)
	mainRepoPath := filepath.Dir(mainGitDir)

	return mainRepoPath, nil
}

// BranchExists checks if a branch exists in the repository.
func BranchExists(mainRepoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", mainRepoPath, "rev-parse", "--verify", branchName)
	err := cmd.Run()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 128 {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// IsBranchCheckedOut checks if a branch is currently checked out in any worktree.
func IsBranchCheckedOut(mainRepoPath, branchName string) (checkedOut bool, location string, err error) {
	worktrees, err := ListWorktrees(mainRepoPath)
	if err != nil {
		return false, "", err
	}

	for _, wt := range worktrees {
		if wt.Branch == "refs/heads/"+branchName {
			return true, wt.Path, nil
		}
	}

	return false, "", nil
}

// DeleteBranch force deletes a branch from the repository.
func DeleteBranch(mainRepoPath, branchName string) error {
	cmd := exec.Command("git", "-C", mainRepoPath, "branch", "-D", branchName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if strings.Contains(errorMsg, "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete branch: %s: %w", errorMsg, err)
	}

	return nil
}

// GetCurrentBranch returns the name of the currently checked out branch.
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		return "", fmt.Errorf("repository is in detached HEAD state")
	}

	return branch, nil
}

// HasUnpushedCommits checks if the branch has commits not pushed to origin.
func HasUnpushedCommits(repoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-list", fmt.Sprintf("origin/%s..%s", branchName, branchName))
	output, err := cmd.Output()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 128 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check for unpushed commits: %w", err)
	}

	commits := strings.TrimSpace(string(output))
	return commits != "", nil
}

// GetUnpushedCommitCount returns the number of commits not pushed to origin.
func GetUnpushedCommitCount(repoPath, branchName string) (int, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", fmt.Sprintf("origin/%s..%s", branchName, branchName))
	output, err := cmd.Output()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 128 {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to count unpushed commits: %w", err)
	}

	countStr := strings.TrimSpace(string(output))
	if countStr == "" {
		return 0, nil
	}

	var count int
	_, err = fmt.Sscanf(countStr, "%d", &count)
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count, nil
}

// GetAllBranches returns a list of all local branch names in the repository.
func GetAllBranches(mainRepoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", mainRepoPath, "branch", "--format=%(refname:short)")
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch != "" {
			result = append(result, branch)
		}
	}

	return result, nil
}

// GetBranchLastCommit returns the relative time and author of the last commit on a branch.
func GetBranchLastCommit(mainRepoPath, branchName string) (relativeTime, author string, err error) {
	cmd := exec.Command("git", "-C", mainRepoPath, "log", "-1", "--format=%cr|%an", branchName)
	output, err := cmd.Output()

	if err != nil {
		return "", "", fmt.Errorf("failed to get branch info: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) != 2 {
		return "", "", nil
	}

	return parts[0], parts[1], nil
}

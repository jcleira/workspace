// Package git provides git repository operations.
package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneRepository clones a git repository into the specified workspace path.
func CloneRepository(workspacePath, repoURL string) error {
	repoName := filepath.Base(repoURL)
	repoName = strings.TrimSuffix(repoName, ".git")

	cmd := exec.Command("git", "clone", repoURL, filepath.Join(workspacePath, repoName))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// CloneRepositoryWithOutput clones a repository with detailed output and error handling.
func CloneRepositoryWithOutput(workspacePath, repoURL, repoName string) error {
	cmd := exec.Command("git", "clone", repoURL, filepath.Join(workspacePath, repoName))
	output, err := cmd.CombinedOutput()

	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		return fmt.Errorf("failed to clone %s: %s", repoName, errorMsg)
	}

	return nil
}

// CheckoutBranch checks out a branch in the specified repository.
func CheckoutBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", branchName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %s: %w", branchName, strings.TrimSpace(string(output)), err)
	}

	return nil
}

// FetchRemote fetches the latest changes from the origin remote.
func FetchRemote(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "origin")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to fetch from origin: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// GetDefaultBranch returns the default branch name for the repository.
func GetDefaultBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()

	if err != nil {
		cmd = exec.Command("git", "-C", repoPath, "remote", "show", "origin")
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to determine default branch: %w", err)
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "HEAD branch:") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}

		return "main", nil
	}

	branch := strings.TrimSpace(string(output))
	branch = strings.TrimPrefix(branch, "refs/remotes/origin/")

	return branch, nil
}

// PullDefaultBranch pulls the latest changes from the default branch.
func PullDefaultBranch(repoPath string) error {
	currentBranch, err := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(currentBranch))

	defaultBranch, err := GetDefaultBranch(repoPath)
	if err != nil {
		return err
	}

	if branch != defaultBranch {
		if checkoutErr := CheckoutBranch(repoPath, defaultBranch); checkoutErr != nil {
			return fmt.Errorf("failed to checkout %s: %w", defaultBranch, checkoutErr)
		}
	}

	cmd := exec.Command("git", "-C", repoPath, "pull", "origin", defaultBranch)
	output, err := cmd.CombinedOutput()

	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if strings.Contains(errorMsg, "Already up to date") || strings.Contains(errorMsg, "Already up-to-date") {
			return nil
		}
		return fmt.Errorf("failed to pull %s: %s: %w", defaultBranch, errorMsg, err)
	}

	return nil
}

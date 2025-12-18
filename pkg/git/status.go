package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// RepoStatusInfo contains comprehensive status information about a git repository.
type RepoStatusInfo struct {
	Branch           string
	HasUncommitted   bool
	UncommittedCount int
	AheadCount       int
	BehindCount      int
	HasUpstream      bool
	LastFetch        time.Time
	LastCommit       string
	LastAuthor       string
	Error            error
}

// HasUncommittedChanges checks if the repository has uncommitted changes.
func HasUncommittedChanges(repoPath string) (hasChanges bool, count int, err error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()

	if err != nil {
		return false, 0, fmt.Errorf("failed to check uncommitted changes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return false, 0, nil
	}

	return true, len(lines), nil
}

// GetAheadBehind returns how many commits the branch is ahead/behind its upstream.
func GetAheadBehind(repoPath string) (ahead, behind int, hasUpstream bool, err error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	output, err := cmd.Output()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return 0, 0, false, nil
		}
		return 0, 0, false, fmt.Errorf("failed to get ahead/behind count: %w", err)
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0, false, nil
	}

	behind, errBehind := strconv.Atoi(parts[0])
	ahead, errAhead := strconv.Atoi(parts[1])
	if errBehind != nil || errAhead != nil {
		return 0, 0, false, fmt.Errorf("failed to parse ahead/behind counts")
	}

	return ahead, behind, true, nil
}

// GetLastFetchTime returns when the repository was last fetched.
func GetLastFetchTime(repoPath string) (time.Time, error) {
	fetchHeadPath := filepath.Join(repoPath, ".git", "FETCH_HEAD")

	if info, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil && !info.IsDir() {
		gitFilePath := filepath.Join(repoPath, ".git")
		content, err := os.ReadFile(gitFilePath)
		if err == nil && strings.HasPrefix(string(content), "gitdir:") {
			gitdir := strings.TrimSpace(strings.TrimPrefix(string(content), "gitdir:"))
			fetchHeadPath = filepath.Join(gitdir, "FETCH_HEAD")
		}
	}

	info, err := os.Stat(fetchHeadPath)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to get fetch time: %w", err)
	}

	return info.ModTime(), nil
}

// GetModifiedFiles returns a list of modified file paths.
func GetModifiedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to get modified files: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, nil
	}

	lines := strings.Split(outputStr, "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		if len(line) > 3 {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}

	return files, nil
}

// GetFullRepoStatus gathers all status info for a repository.
func GetFullRepoStatus(repoPath string) RepoStatusInfo {
	status := RepoStatusInfo{}

	branch, err := GetCurrentBranch(repoPath)
	if err != nil {
		status.Error = err
		return status
	}
	status.Branch = branch

	hasUncommitted, count, err := HasUncommittedChanges(repoPath)
	if err != nil {
		status.Error = err
		return status
	}
	status.HasUncommitted = hasUncommitted
	status.UncommittedCount = count

	ahead, behind, hasUpstream, err := GetAheadBehind(repoPath)
	if err != nil {
		status.Error = err
		return status
	}
	status.AheadCount = ahead
	status.BehindCount = behind
	status.HasUpstream = hasUpstream

	if lastFetch, err := GetLastFetchTime(repoPath); err == nil {
		status.LastFetch = lastFetch
	}

	if lastCommit, lastAuthor, err := GetBranchLastCommit(repoPath, branch); err == nil {
		status.LastCommit = lastCommit
		status.LastAuthor = lastAuthor
	}

	return status
}

// NeedsFetch checks if the repository needs to be fetched.
func NeedsFetch(repoPath string, threshold time.Duration) (bool, time.Duration, error) {
	lastFetch, err := GetLastFetchTime(repoPath)
	if err != nil {
		return false, 0, err
	}

	if lastFetch.IsZero() {
		return true, 0, nil
	}

	sinceLastFetch := time.Since(lastFetch)
	return sinceLastFetch > threshold, sinceLastFetch, nil
}

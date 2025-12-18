package status

import (
	"path/filepath"

	"github.com/jcleira/workspace/pkg/git"
)

// GetRepoStatus gathers all status info for a single repository.
func GetRepoStatus(repoPath, name string) RepoStatus {
	status := RepoStatus{
		Name: name,
		Path: repoPath,
	}

	branch, err := git.GetCurrentBranch(repoPath)
	if err != nil {
		status.Error = err
		return status
	}
	status.Branch = branch

	hasUncommitted, count, err := git.HasUncommittedChanges(repoPath)
	if err != nil {
		status.Error = err
		return status
	}
	status.HasUncommitted = hasUncommitted
	status.UncommittedCount = count

	unpushedCount, err := git.GetUnpushedCommitCount(repoPath, branch)
	if err == nil {
		status.UnpushedCount = unpushedCount
	}

	ahead, behind, hasUpstream, err := git.GetAheadBehind(repoPath)
	if err == nil {
		status.AheadCount = ahead
		status.BehindCount = behind
		status.HasUpstream = hasUpstream
	}

	if lastFetch, err := git.GetLastFetchTime(repoPath); err == nil {
		status.LastFetch = lastFetch
	}

	if lastCommit, lastAuthor, err := git.GetBranchLastCommit(repoPath, branch); err == nil {
		status.LastCommitTime = lastCommit
		status.LastCommitAuthor = lastAuthor
	}

	return status
}

// GetWorkspaceStatus fetches status for all repos in a workspace.
func GetWorkspaceStatus(workspacePath string, projects []string) []RepoStatus {
	statuses := make([]RepoStatus, 0, len(projects))

	for _, project := range projects {
		repoPath := filepath.Join(workspacePath, project)
		status := GetRepoStatus(repoPath, project)
		statuses = append(statuses, status)
	}

	return statuses
}

// CountIssues counts how many repos have uncommitted/unpushed changes.
func CountIssues(statuses []RepoStatus) int {
	count := 0
	for i := range statuses {
		if statuses[i].HasUncommitted || statuses[i].UnpushedCount > 0 || statuses[i].AheadCount > 0 {
			count++
		}
	}
	return count
}

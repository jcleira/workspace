package branch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jcleira/workspace/pkg/git"
	"github.com/jcleira/workspace/pkg/workspace"
)

// Service provides branch management operations.
type Service struct {
	workspaceManager *workspace.Manager
	ignorePatterns   []string
}

// NewService creates a new branch service.
func NewService(wm *workspace.Manager, ignorePatterns []string) *Service {
	return &Service{
		workspaceManager: wm,
		ignorePatterns:   ignorePatterns,
	}
}

// BuildWorkspaceMapping creates a map of repo:branch -> workspace name.
func (s *Service) BuildWorkspaceMapping() (WorkspaceBranchMapping, error) {
	workspaces, err := s.workspaceManager.GetWorkspaces()
	if err != nil {
		return nil, err
	}

	mapping := make(WorkspaceBranchMapping)
	for _, ws := range workspaces {
		wsName := strings.TrimPrefix(ws.Name, "workspace-")
		if wsName == "default" {
			continue
		}

		entries, err := os.ReadDir(ws.Path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == ".claude" {
				continue
			}

			repoPath := filepath.Join(ws.Path, entry.Name())
			isWorktree, err := git.IsWorktree(repoPath)
			if err != nil || !isWorktree {
				continue
			}

			branch, err := git.GetCurrentBranch(repoPath)
			if err == nil {
				key := fmt.Sprintf("%s:%s", entry.Name(), branch)
				mapping[key] = wsName
			}
		}
	}

	return mapping, nil
}

// List returns all branches across repositories with workspace mapping.
func (s *Service) List() (*ListOutput, error) {
	reposDir := s.workspaceManager.ReposDir
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return &ListOutput{}, nil
	}

	repos, err := os.ReadDir(reposDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read repositories: %w", err)
	}

	workspaceMapping, err := s.BuildWorkspaceMapping()
	if err != nil {
		return nil, fmt.Errorf("failed to build workspace mapping: %w", err)
	}

	output := &ListOutput{
		Repositories: make([]RepositoryBranches, 0),
	}

	for _, repo := range repos {
		if !repo.IsDir() {
			continue
		}

		repoPath := filepath.Join(reposDir, repo.Name())
		branches, err := git.GetAllBranches(repoPath)
		if err != nil {
			continue
		}

		defaultBranch, err := git.GetDefaultBranch(repoPath)
		if err != nil {
			defaultBranch = ""
		}

		repoBranches := RepositoryBranches{
			RepoName:      repo.Name(),
			RepoPath:      repoPath,
			DefaultBranch: defaultBranch,
			Branches:      make([]BranchInfo, 0, len(branches)),
		}

		for _, branch := range branches {
			output.TotalBranches++

			info := BranchInfo{
				Name:      branch,
				RepoName:  repo.Name(),
				RepoPath:  repoPath,
				IsDefault: branch == defaultBranch,
				IsIgnored: git.ShouldIgnoreBranch(branch, s.ignorePatterns),
			}

			if info.IsIgnored {
				output.IgnoredCount++
			}

			key := fmt.Sprintf("%s:%s", repo.Name(), branch)
			if wsName, ok := workspaceMapping[key]; ok {
				info.WorkspaceName = wsName
			} else if !info.IsDefault && !info.IsIgnored {
				info.IsOrphaned = true
				output.OrphanedCount++
			}

			hasUnpushed, err := git.HasUnpushedCommits(repoPath, branch)
			if err == nil && hasUnpushed {
				info.HasUnpushed = true
				if count, err := git.GetUnpushedCommitCount(repoPath, branch); err == nil {
					info.UnpushedCount = count
				}
			}

			if commitTime, commitBy, err := git.GetBranchLastCommit(repoPath, branch); err == nil {
				info.LastCommitTime = commitTime
				info.LastCommitBy = commitBy
			}

			repoBranches.Branches = append(repoBranches.Branches, info)
		}

		output.Repositories = append(output.Repositories, repoBranches)
	}

	return output, nil
}

// PlanCleanup identifies orphaned branches for cleanup.
func (s *Service) PlanCleanup() (*CleanupPlan, error) {
	listOutput, err := s.List()
	if err != nil {
		return nil, err
	}

	plan := &CleanupPlan{
		OrphanedBranches: make([]OrphanedBranch, 0),
		SkippedIgnored:   listOutput.IgnoredCount,
	}

	for _, repo := range listOutput.Repositories {
		for _, branch := range repo.Branches {
			if branch.IsOrphaned {
				plan.OrphanedBranches = append(plan.OrphanedBranches, OrphanedBranch{
					RepoName:      branch.RepoName,
					RepoPath:      branch.RepoPath,
					BranchName:    branch.Name,
					HasUnpushed:   branch.HasUnpushed,
					UnpushedCount: branch.UnpushedCount,
				})
			}
		}
	}

	return plan, nil
}

// ExecuteCleanup deletes the specified orphaned branches.
// skipBranches contains branch names to skip (format: "repo:branch").
func (s *Service) ExecuteCleanup(plan *CleanupPlan, skipBranches []string) (*CleanupResult, error) {
	skipSet := make(map[string]bool)
	for _, skip := range skipBranches {
		skipSet[skip] = true
	}

	result := &CleanupResult{
		Deleted: make([]BranchDeleteResult, 0),
		Skipped: make([]BranchDeleteResult, 0),
		Failed:  make([]BranchDeleteResult, 0),
	}

	for _, ob := range plan.OrphanedBranches {
		key := fmt.Sprintf("%s:%s", ob.RepoName, ob.BranchName)
		if skipSet[key] {
			result.Skipped = append(result.Skipped, BranchDeleteResult{
				RepoName:   ob.RepoName,
				BranchName: ob.BranchName,
				Reason:     "user skipped",
			})
			continue
		}

		if err := git.DeleteBranch(ob.RepoPath, ob.BranchName); err != nil {
			result.Failed = append(result.Failed, BranchDeleteResult{
				RepoName:   ob.RepoName,
				BranchName: ob.BranchName,
				Error:      err,
			})
		} else {
			result.Deleted = append(result.Deleted, BranchDeleteResult{
				RepoName:   ob.RepoName,
				BranchName: ob.BranchName,
			})
		}
	}

	return result, nil
}

// DeleteBranch deletes a single branch from a repository.
func (s *Service) DeleteBranch(repoPath, branchName string) error {
	return git.DeleteBranch(repoPath, branchName)
}

// CheckUnpushed checks if a branch has unpushed commits.
func (s *Service) CheckUnpushed(repoPath, branchName string) (hasUnpushed bool, count int, err error) {
	hasUnpushed, err = git.HasUnpushedCommits(repoPath, branchName)
	if err != nil {
		return false, 0, err
	}
	if hasUnpushed {
		count, err = git.GetUnpushedCommitCount(repoPath, branchName)
		if err != nil {
			return hasUnpushed, 0, err
		}
	}
	return hasUnpushed, count, nil
}

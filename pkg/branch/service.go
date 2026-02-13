package branch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

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
func (s *Service) List() (ListOutput, error) {
	reposDir := s.workspaceManager.ReposDir
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return ListOutput{}, nil
	}

	repos, err := os.ReadDir(reposDir)
	if err != nil {
		return ListOutput{}, fmt.Errorf("failed to read repositories: %w", err)
	}

	workspaceMapping, err := s.BuildWorkspaceMapping()
	if err != nil {
		return ListOutput{}, fmt.Errorf("failed to build workspace mapping: %w", err)
	}

	output := ListOutput{
		Repositories: make([]RepositoryBranches, 0, len(repos)),
	}

	var mu sync.Mutex
	var g errgroup.Group

	for _, repo := range repos {
		if !repo.IsDir() {
			continue
		}

		repoName := repo.Name()
		repoPath := filepath.Join(reposDir, repoName)

		g.Go(func() error {
			branches, err := git.GetAllBranches(repoPath)
			if err != nil {
				return err
			}

			defaultBranch, err := git.GetDefaultBranch(repoPath)
			if err != nil {
				defaultBranch = ""
			}

			repoBranches := RepositoryBranches{
				RepoName:      repoName,
				RepoPath:      repoPath,
				DefaultBranch: defaultBranch,
				Branches:      make([]BranchInfo, 0, len(branches)),
			}

			var branchMu sync.Mutex
			var branchGroup errgroup.Group

			localTotalBranches := 0
			localIgnoredCount := 0
			localOrphanedCount := 0

			for _, branch := range branches {
				branchName := branch
				branchGroup.Go(func() error {
					info := BranchInfo{
						Name:      branchName,
						RepoName:  repoName,
						RepoPath:  repoPath,
						IsDefault: branchName == defaultBranch,
						IsIgnored: git.ShouldIgnoreBranch(branchName, s.ignorePatterns),
					}

					key := fmt.Sprintf("%s:%s", repoName, branchName)
					if wsName, ok := workspaceMapping[key]; ok {
						info.WorkspaceName = wsName
					} else if !info.IsDefault && !info.IsIgnored {
						info.IsOrphaned = true
					}

					hasUnpushed, err := git.HasUnpushedCommits(repoPath, branchName)
					if err == nil && hasUnpushed {
						info.HasUnpushed = true
						if count, err := git.GetUnpushedCommitCount(repoPath, branchName); err == nil {
							info.UnpushedCount = count
						}
					}

					if commitTime, commitBy, err := git.GetBranchLastCommit(repoPath, branchName); err == nil {
						info.LastCommitTime = commitTime
						info.LastCommitBy = commitBy
					}

					branchMu.Lock()
					repoBranches.Branches = append(repoBranches.Branches, info)
					localTotalBranches++
					if info.IsIgnored {
						localIgnoredCount++
					}
					if info.IsOrphaned {
						localOrphanedCount++
					}
					branchMu.Unlock()
					return nil
				})
			}

			if err := branchGroup.Wait(); err != nil {
				return err
			}

			mu.Lock()
			output.Repositories = append(output.Repositories, repoBranches)
			output.TotalBranches += localTotalBranches
			output.IgnoredCount += localIgnoredCount
			output.OrphanedCount += localOrphanedCount
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return ListOutput{}, err
	}
	return output, nil
}

// PlanCleanup identifies orphaned branches for cleanup.
func (s *Service) PlanCleanup() (CleanupPlan, error) {
	listOutput, err := s.List()
	if err != nil {
		return CleanupPlan{}, err
	}

	plan := CleanupPlan{
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

// ExecuteCleanup deletes the specified orphaned branches in parallel.
// skipBranches contains branch names to skip (format: "repo:branch").
func (s *Service) ExecuteCleanup(plan CleanupPlan, skipBranches []string) (CleanupResult, error) {
	skipSet := make(map[string]bool)
	for _, skip := range skipBranches {
		skipSet[skip] = true
	}

	var deleted, skipped, failed []BranchDeleteResult
	var mu sync.Mutex
	var g errgroup.Group

	for _, ob := range plan.OrphanedBranches {
		key := fmt.Sprintf("%s:%s", ob.RepoName, ob.BranchName)
		if skipSet[key] {
			skipped = append(skipped, BranchDeleteResult{
				RepoName:   ob.RepoName,
				BranchName: ob.BranchName,
				Reason:     "user skipped",
			})
			continue
		}

		g.Go(func() error {
			if err := git.DeleteBranch(ob.RepoPath, ob.BranchName); err != nil {
				mu.Lock()
				failed = append(failed, BranchDeleteResult{
					RepoName:   ob.RepoName,
					BranchName: ob.BranchName,
					Error:      err,
				})
				mu.Unlock()
			} else {
				mu.Lock()
				deleted = append(deleted, BranchDeleteResult{
					RepoName:   ob.RepoName,
					BranchName: ob.BranchName,
				})
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return CleanupResult{}, err
	}
	return CleanupResult{
		Deleted: deleted,
		Skipped: skipped,
		Failed:  failed,
	}, nil
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

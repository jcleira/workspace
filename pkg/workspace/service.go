package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jcleira/workspace/pkg/git"
)

// Service provides workspace management operations.
type Service struct {
	manager *Manager
}

// NewService creates a new workspace service.
func NewService(manager *Manager) *Service {
	return &Service{
		manager: manager,
	}
}

// SyncMainRepos fetches and pulls all main repositories.
func (s *Service) SyncMainRepos(repos []RepositorySpec) []SyncResult {
	results := make([]SyncResult, 0, len(repos))

	for _, repo := range repos {
		mainRepoPath := filepath.Join(s.manager.ReposDir, repo.Name)

		result := SyncResult{
			RepoName: repo.Name,
			RepoPath: mainRepoPath,
		}

		if _, err := os.Stat(mainRepoPath); os.IsNotExist(err) {
			continue
		}

		if err := git.FetchRemote(mainRepoPath); err != nil {
			result.Error = fmt.Errorf("failed to fetch: %w", err)
			results = append(results, result)
			continue
		}
		result.Fetched = true

		if err := git.PullDefaultBranch(mainRepoPath); err != nil {
			result.Error = fmt.Errorf("failed to pull: %w", err)
			results = append(results, result)
			continue
		}
		result.Pulled = true

		results = append(results, result)
	}

	return results
}

// Create creates a new workspace with repositories.
func (s *Service) Create(input CreateInput) (*CreateOutput, error) {
	repos, err := DiscoverMainRepos(s.manager.ReposDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found in %s", s.manager.ReposDir)
	}

	useWorktrees := input.Name != "default"

	output := &CreateOutput{
		WorkspaceType: WorkspaceTypeClone,
		CreatedRepos:  make([]RepoResult, 0),
		FailedRepos:   make([]RepoResult, 0),
		SyncResults:   make([]SyncResult, 0),
	}

	if useWorktrees {
		output.WorkspaceType = WorkspaceTypeWorktree
		output.SyncResults = s.SyncMainRepos(repos)
	}

	workspacePath, alreadyExists, err := s.createWorkspaceDir(input.Name)
	if err != nil {
		return nil, err
	}
	output.WorkspacePath = workspacePath
	output.AlreadyExists = alreadyExists

	for _, repo := range repos {
		targetPath := filepath.Join(workspacePath, repo.Name)

		if useWorktrees {
			result := s.createWorktree(repo, targetPath, input.Name)
			if result.Error != nil {
				output.FailedRepos = append(output.FailedRepos, result)
			} else {
				output.CreatedRepos = append(output.CreatedRepos, result)
			}
		} else {
			result := s.cloneRepo(repo, workspacePath)
			if result.Error != nil {
				output.FailedRepos = append(output.FailedRepos, result)
			} else {
				output.CreatedRepos = append(output.CreatedRepos, result)
			}
		}
	}

	if err := CreateWorkspaceInfo(workspacePath, input.Name, "", len(output.CreatedRepos), len(output.FailedRepos)); err != nil {
		return nil, fmt.Errorf("failed to write workspace info: %w", err)
	}

	return output, nil
}

func (s *Service) createWorkspaceDir(name string) (workspacePath string, alreadyExists bool, err error) {
	workspacePath = filepath.Join(s.manager.WorkspacesDir, "workspace-"+name)
	alreadyExists = false

	if _, err := os.Stat(workspacePath); err == nil {
		alreadyExists = true
	} else {
		if err := os.MkdirAll(workspacePath, 0o755); err != nil {
			return "", false, fmt.Errorf("failed to create workspace: %w", err)
		}
	}

	claudeSymlink := filepath.Join(workspacePath, ".claude")
	if _, err := os.Lstat(claudeSymlink); os.IsNotExist(err) {
		if err := os.Symlink(s.manager.ClaudeDir, claudeSymlink); err != nil {
			return "", false, fmt.Errorf("failed to create claude symlink: %w", err)
		}
	}

	return workspacePath, alreadyExists, nil
}

func (s *Service) createWorktree(repo RepositorySpec, targetPath, branchName string) RepoResult {
	result := RepoResult{
		Name:       repo.Name,
		Path:       targetPath,
		BranchName: branchName,
	}

	mainRepoPath := filepath.Join(s.manager.ReposDir, repo.Name)

	if _, err := os.Stat(mainRepoPath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("main repository not found")
		return result
	}

	exists, err := git.BranchExists(mainRepoPath, branchName)
	if err != nil {
		result.Error = fmt.Errorf("could not check branch existence: %w", err)
		return result
	}

	if exists {
		checkedOut, location, err := git.IsBranchCheckedOut(mainRepoPath, branchName)
		if err != nil {
			result.Error = fmt.Errorf("could not check if branch is checked out: %w", err)
			return result
		}
		if checkedOut {
			result.Error = fmt.Errorf("branch '%s' is already checked out at %s", branchName, location)
			return result
		}

		result.WasExisting = true
		if err := git.CheckoutExistingBranch(mainRepoPath, targetPath, branchName); err != nil {
			result.Error = fmt.Errorf("failed to checkout existing branch: %w", err)
			return result
		}
	} else {
		if err := git.CreateWorktree(mainRepoPath, targetPath, branchName); err != nil {
			result.Error = fmt.Errorf("failed to create worktree: %w", err)
			return result
		}
	}

	return result
}

func (s *Service) cloneRepo(repo RepositorySpec, workspacePath string) RepoResult {
	result := RepoResult{
		Name: repo.Name,
		Path: filepath.Join(workspacePath, repo.Name),
	}

	if err := git.CloneRepositoryWithOutput(workspacePath, repo.URL, repo.Name); err != nil {
		result.Error = fmt.Errorf("failed to clone: %w", err)
	}

	return result
}

// Delete removes a workspace and optionally its associated branches.
func (s *Service) Delete(input DeleteInput) (*DeleteOutput, error) {
	if input.Name == "default" {
		return nil, fmt.Errorf("the 'default' workspace is protected and cannot be deleted")
	}

	workspacePath := filepath.Join(s.manager.WorkspacesDir, "workspace-"+input.Name)

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace 'workspace-%s' does not exist", input.Name)
	}

	output := &DeleteOutput{
		WorkspacePath:   workspacePath,
		DeletedBranches: make([]BranchResult, 0),
		SkippedBranches: make([]BranchResult, 0),
	}

	wsType, err := s.manager.GetWorkspaceType(workspacePath)
	if err == nil {
		output.WorkspaceType = wsType
	}

	if wsType == WorkspaceTypeWorktree {
		s.cleanupWorktrees(workspacePath, input.DeleteBranches, output)
	}

	if err := os.RemoveAll(workspacePath); err != nil {
		return nil, fmt.Errorf("failed to delete workspace: %w", err)
	}

	output.Deleted = true
	return output, nil
}

func (s *Service) cleanupWorktrees(workspacePath string, deleteBranches bool, output *DeleteOutput) {
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".claude" {
			continue
		}

		repoPath := filepath.Join(workspacePath, entry.Name())
		isWorktree, err := git.IsWorktree(repoPath)
		if err != nil || !isWorktree {
			continue
		}

		mainRepoPath, err := git.GetMainRepoFromWorktree(repoPath)
		if err != nil {
			continue
		}

		var branchName string
		if deleteBranches {
			branchName, err = git.GetCurrentBranch(repoPath)
			if err != nil {
				branchName = ""
			}
		}

		if err := git.RemoveWorktree(mainRepoPath, repoPath); err != nil {
			continue
		}

		if deleteBranches && branchName != "" {
			result := BranchResult{
				RepoName:   entry.Name(),
				BranchName: branchName,
			}

			if err := git.DeleteBranch(mainRepoPath, branchName); err != nil {
				result.Error = err
				output.SkippedBranches = append(output.SkippedBranches, result)
			} else {
				output.DeletedBranches = append(output.DeletedBranches, result)
			}
		}
	}
}

// List returns all workspaces.
func (s *Service) List() ([]WorkspaceInfo, error) {
	workspaces, err := s.manager.GetWorkspaces()
	if err != nil {
		return nil, err
	}

	infos := make([]WorkspaceInfo, len(workspaces))
	for i, ws := range workspaces {
		infos[i] = ToWorkspaceInfo(ws)
	}

	return infos, nil
}

// GetPath returns the filesystem path for a workspace.
func (s *Service) GetPath(name string) (string, error) {
	return s.manager.GetWorkspacePath(name)
}

// GetType determines if a workspace uses clones or worktrees.
func (s *Service) GetType(workspacePath string) (WorkspaceType, error) {
	return s.manager.GetWorkspaceType(workspacePath)
}

// DiscoverRepos finds all git repositories in the repos directory.
func (s *Service) DiscoverRepos() ([]RepositorySpec, error) {
	return DiscoverMainRepos(s.manager.ReposDir)
}

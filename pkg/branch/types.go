// Package branch provides branch management types and services.
package branch

// BranchInfo contains comprehensive information about a branch.
type BranchInfo struct {
	Name           string
	RepoName       string
	RepoPath       string
	IsDefault      bool
	WorkspaceName  string
	IsOrphaned     bool
	IsIgnored      bool
	HasUnpushed    bool
	UnpushedCount  int
	LastCommitTime string
	LastCommitBy   string
}

// ListInput contains parameters for listing branches.
type ListInput struct {
	ReposDir       string
	WorkspacesDir  string
	IgnorePatterns []string
}

// ListOutput contains the result of listing branches.
type ListOutput struct {
	Repositories  []RepositoryBranches
	TotalBranches int
	OrphanedCount int
	IgnoredCount  int
}

// RepositoryBranches groups branches by repository.
type RepositoryBranches struct {
	RepoName      string
	RepoPath      string
	DefaultBranch string
	Branches      []BranchInfo
}

// CleanupInput contains parameters for cleanup operation.
type CleanupInput struct {
	ReposDir       string
	IgnorePatterns []string
}

// CleanupPlan describes what branches will be cleaned up.
type CleanupPlan struct {
	OrphanedBranches []OrphanedBranch
	SkippedIgnored   int
}

// OrphanedBranch is a branch that can be cleaned up.
type OrphanedBranch struct {
	RepoName      string
	RepoPath      string
	BranchName    string
	HasUnpushed   bool
	UnpushedCount int
}

// CleanupResult contains the outcome of cleanup operation.
type CleanupResult struct {
	Deleted []BranchDeleteResult
	Skipped []BranchDeleteResult
	Failed  []BranchDeleteResult
}

// BranchDeleteResult describes outcome of deleting a single branch.
type BranchDeleteResult struct {
	RepoName   string
	BranchName string
	Reason     string
	Error      error
}

// WorkspaceBranchMapping maps repo:branch to workspace name.
type WorkspaceBranchMapping map[string]string

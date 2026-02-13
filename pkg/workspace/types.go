package workspace

import "time"

// ProgressCallback is called to report progress during operations.
type ProgressCallback func(message string)

// CreateInput contains all parameters for creating a workspace.
type CreateInput struct {
	Name       string
	OnProgress ProgressCallback
}

// CreateOutput contains the result of workspace creation.
type CreateOutput struct {
	WorkspacePath string
	CreatedRepos  []RepoResult
	FailedRepos   []RepoResult
	SyncResults   []SyncResult
	WorkspaceType WorkspaceType
	AlreadyExists bool
}

// RepoResult describes the outcome of a single repository operation.
type RepoResult struct {
	Name        string
	Path        string
	BranchName  string
	Error       error
	WasExisting bool
}

// SyncResult describes the outcome of syncing a main repository.
type SyncResult struct {
	RepoName string
	RepoPath string
	Fetched  bool
	Pulled   bool
	Error    error
}

// DeleteInput contains all parameters for deleting a workspace.
type DeleteInput struct {
	Name           string
	DeleteBranches bool
}

// DeleteOutput contains the result of workspace deletion.
type DeleteOutput struct {
	WorkspacePath   string
	WorkspaceType   WorkspaceType
	DeletedBranches []BranchResult
	SkippedBranches []BranchResult
	Deleted         bool
}

// BranchResult describes the outcome of a branch operation.
type BranchResult struct {
	RepoName      string
	BranchName    string
	UnpushedCount int
	Error         error
}

// WorkspaceInfo extends the basic Workspace with additional metadata.
type WorkspaceInfo struct {
	Name      string
	Path      string
	Projects  []string
	Created   time.Time
	Type      WorkspaceType
	IsDefault bool
}

// ToWorkspaceInfo converts a Workspace to WorkspaceInfo.
func ToWorkspaceInfo(w Workspace) WorkspaceInfo {
	return WorkspaceInfo{
		Name:      w.Name,
		Path:      w.Path,
		Projects:  w.Projects,
		Created:   w.Created,
		IsDefault: w.Name == DefaultWorkspaceName,
	}
}

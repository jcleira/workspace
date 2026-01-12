package dashboard

import (
	"time"

	"github.com/jcleira/workspace/pkg/workspace"
)

// TickMsg is sent periodically for auto-refresh.
type TickMsg time.Time

// RefreshMsg requests a refresh of all workspace statuses.
type RefreshMsg struct{}

// WorkspacesLoadedMsg is sent when workspaces are loaded.
type WorkspacesLoadedMsg struct {
	Workspaces []WorkspaceData
}

// WorkspaceStatusMsg is sent when a workspace's status is updated.
type WorkspaceStatusMsg struct {
	WorkspacePath string
	Statuses      []RepoStatus
}

// RepoStatusMsg is sent when a single repo's status is updated.
type RepoStatusMsg struct {
	WorkspacePath string
	RepoName      string
	Status        RepoStatus
}

// ActionStartMsg is sent when an action starts.
type ActionStartMsg struct {
	Action    ActionType
	Workspace string
	Repo      string
}

// ActionCompleteMsg is sent when an action completes.
type ActionCompleteMsg struct {
	Action    ActionType
	Workspace string
	Repo      string
	Success   bool
	Error     error
	Output    string
}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Error   error
	Context string
}

// DiffLoadedMsg is sent when diff content is loaded.
type DiffLoadedMsg struct {
	Content  string
	RepoName string
	Staged   bool
	Error    error
}

// WorkspaceData contains information about a workspace.
type WorkspaceData struct {
	Name       string
	Path       string
	Projects   []string
	Created    time.Time
	IsWorktree bool
}

// ActionType represents the type of action to perform.
type ActionType int

const (
	ActionFetch ActionType = iota
	ActionPull
	ActionDelete
	ActionCreate
	ActionSwitch
	ActionDiff
)

// String returns a human-readable name for the action.
func (a ActionType) String() string {
	switch a {
	case ActionFetch:
		return "Fetch"
	case ActionPull:
		return "Pull"
	case ActionDelete:
		return "Delete"
	case ActionCreate:
		return "Create"
	case ActionSwitch:
		return "Switch"
	case ActionDiff:
		return "Diff"
	default:
		return "Unknown"
	}
}

// ConfirmAction contains details for a confirmation dialog.
type ConfirmAction struct {
	Action      ActionType
	Workspace   string
	Repo        string
	Description string
}

// WorkspacesToData converts workspace types to dashboard WorkspaceData.
func WorkspacesToData(workspaces []workspace.Workspace) []WorkspaceData {
	data := make([]WorkspaceData, len(workspaces))
	for i, ws := range workspaces {
		data[i] = WorkspaceData{
			Name:     ws.Name,
			Path:     ws.Path,
			Projects: ws.Projects,
			Created:  ws.Created,
		}
	}
	return data
}

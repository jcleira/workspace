package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	claudeDirName = ".claude"

	// DefaultWorkspaceName is the name of the protected default workspace.
	DefaultWorkspaceName = "workspace-default"
)

// Manager handles workspace operations including creation, deletion, and listing.
type Manager struct {
	WorkspacesDir string
	ReposDir      string
	ClaudeDir     string
}

// NewManager creates a new workspace manager with the specified directories.
func NewManager(workspacesDir, reposDir, claudeDir string) *Manager {
	return &Manager{
		WorkspacesDir: workspacesDir,
		ReposDir:      reposDir,
		ClaudeDir:     claudeDir,
	}
}

// GetWorkspaces returns a list of all workspaces sorted by creation time.
func (m *Manager) GetWorkspaces() ([]Workspace, error) {
	if _, err := os.Stat(m.WorkspacesDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(m.WorkspacesDir)
	if err != nil {
		return nil, err
	}

	workspaces := make([]Workspace, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "workspace-") {
			continue
		}

		workspacePath := filepath.Join(m.WorkspacesDir, entry.Name())

		ws := Workspace{
			Name: entry.Name(),
			Path: workspacePath,
		}

		if created, err := ReadWorkspaceInfo(workspacePath); err == nil {
			ws.Created = created
		}

		if wsEntries, err := os.ReadDir(workspacePath); err == nil {
			for _, wsEntry := range wsEntries {
				if wsEntry.IsDir() && wsEntry.Name() != claudeDirName {
					entryPath := filepath.Join(workspacePath, wsEntry.Name())
					if _, err := os.Stat(filepath.Join(entryPath, ".git")); err == nil {
						ws.Projects = append(ws.Projects, wsEntry.Name())
					}
				}
			}
		}

		workspaces = append(workspaces, ws)
	}

	sort.Slice(workspaces, func(i, j int) bool {
		if workspaces[i].Name == DefaultWorkspaceName {
			return true
		}
		if workspaces[j].Name == DefaultWorkspaceName {
			return false
		}
		return workspaces[i].Created.After(workspaces[j].Created)
	})

	return workspaces, nil
}

// GetWorkspacePath returns the filesystem path for a workspace by name.
func (m *Manager) GetWorkspacePath(name string) (string, error) {
	workspacePath := filepath.Join(m.WorkspacesDir, "workspace-"+name)

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace 'workspace-%s' does not exist", name)
	}

	return workspacePath, nil
}

// RepositorySpec describes a repository to be included in a workspace.
type RepositorySpec struct {
	URL    string
	Name   string
	Branch string
}

// WorkspaceType indicates whether a workspace uses clones or worktrees.
type WorkspaceType string

const (
	WorkspaceTypeClone    WorkspaceType = "clone"
	WorkspaceTypeWorktree WorkspaceType = "worktree"
)

// GetWorkspaceType determines if a workspace uses clones or worktrees.
func (m *Manager) GetWorkspaceType(workspacePath string) (WorkspaceType, error) {
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != claudeDirName {
			repoPath := filepath.Join(workspacePath, entry.Name())
			gitPath := filepath.Join(repoPath, ".git")

			info, err := os.Stat(gitPath)
			if err == nil {
				if info.IsDir() {
					return WorkspaceTypeClone, nil
				}
				return WorkspaceTypeWorktree, nil
			}
		}
	}

	return "", fmt.Errorf("no git repositories found in workspace")
}

package workspace

import "time"

// Workspace represents a development workspace containing multiple repositories.
// This is a shared type used across packages (shell, ui, cmd).
type Workspace struct {
	Name     string
	Path     string
	Projects []string
	Created  time.Time
}

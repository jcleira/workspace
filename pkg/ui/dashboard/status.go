package dashboard

import (
	"fmt"

	"github.com/jcleira/workspace/pkg/status"
)

// RepoStatus is an alias to the shared status type with UI methods.
type RepoStatus = status.RepoStatus

// StatusSummary is an alias to the shared status summary type.
type StatusSummary = status.StatusSummary

const (
	StatusClean    = status.StatusClean
	StatusModified = status.StatusModified
	StatusAhead    = status.StatusAhead
	StatusBehind   = status.StatusBehind
	StatusDiverged = status.StatusDiverged
	StatusError    = status.StatusError
	StatusLoading  = status.StatusLoading
)

// Icon returns the status icon for display.
func Icon(s RepoStatus) string {
	switch s.Summary() {
	case StatusClean:
		return statusCleanStyle.Render("✓")
	case StatusModified:
		return statusModifiedStyle.Render("*")
	case StatusAhead:
		return statusAheadStyle.Render("^")
	case StatusBehind:
		return statusBehindStyle.Render("v")
	case StatusDiverged:
		return statusWarningStyle.Render("<>")
	case StatusError:
		return statusErrorStyle.Render("!")
	case StatusLoading:
		return dimmedItemStyle.Render("...")
	default:
		return " "
	}
}

var statusWarningStyle = statusModifiedStyle

// StatusText returns a short status description for display.
func StatusText(s RepoStatus) string {
	switch s.Summary() {
	case StatusClean:
		return statusCleanStyle.Render("clean")
	case StatusModified:
		parts := []string{}
		if s.UncommittedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d modified", s.UncommittedCount))
		}
		if s.UnpushedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d unpushed", s.UnpushedCount))
		}
		if s.AheadCount > 0 {
			parts = append(parts, fmt.Sprintf("%d ahead", s.AheadCount))
		}
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += ", "
			}
			result += p
		}
		return statusModifiedStyle.Render(result)
	case StatusAhead:
		return statusAheadStyle.Render(fmt.Sprintf("%d ahead", s.AheadCount))
	case StatusBehind:
		return statusBehindStyle.Render(fmt.Sprintf("%d behind", s.BehindCount))
	case StatusDiverged:
		return statusWarningStyle.Render(fmt.Sprintf("%d ahead, %d behind", s.AheadCount, s.BehindCount))
	case StatusError:
		if s.Error != nil {
			return statusErrorStyle.Render("error: " + s.Error.Error())
		}
		return statusErrorStyle.Render("error")
	case StatusLoading:
		return dimmedItemStyle.Render("loading...")
	default:
		return ""
	}
}

// FetchWorkspaceStatus fetches status for all repos in a workspace.
func FetchWorkspaceStatus(workspacePath string, projects []string) []RepoStatus {
	return status.GetWorkspaceStatus(workspacePath, projects)
}

// CountWorkspaceIssues counts how many repos have uncommitted/unpushed changes.
func CountWorkspaceIssues(statuses []RepoStatus) int {
	return status.CountIssues(statuses)
}

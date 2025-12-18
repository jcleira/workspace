// Package status provides repository status monitoring types and services.
package status

import "time"

// RepoStatus contains the status of a single repository.
type RepoStatus struct {
	Name             string
	Path             string
	Branch           string
	HasUncommitted   bool
	UncommittedCount int
	UnpushedCount    int
	AheadCount       int
	BehindCount      int
	HasUpstream      bool
	LastFetch        time.Time
	LastCommitTime   string
	LastCommitAuthor string
	Error            error
}

// StatusSummary represents the overall status of a repo.
type StatusSummary int

const (
	StatusClean StatusSummary = iota
	StatusModified
	StatusAhead
	StatusBehind
	StatusDiverged
	StatusError
	StatusLoading
)

// Summary returns the overall status summary for a repository.
func (s RepoStatus) Summary() StatusSummary {
	if s.Error != nil {
		return StatusError
	}
	if s.HasUncommitted || s.UnpushedCount > 0 {
		return StatusModified
	}
	if s.AheadCount > 0 && s.BehindCount > 0 {
		return StatusDiverged
	}
	if s.AheadCount > 0 {
		return StatusAhead
	}
	if s.BehindCount > 0 {
		return StatusBehind
	}
	return StatusClean
}

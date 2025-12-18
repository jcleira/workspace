package git

import (
	"path"
	"strings"
)

// MatchesBranchPattern checks if a branch name matches a glob pattern.
func MatchesBranchPattern(branchName, pattern string) bool {
	if pattern == "" || branchName == "" {
		return false
	}

	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		return branchName == pattern
	}

	matched, err := path.Match(pattern, branchName)
	if err != nil {
		return false
	}

	return matched
}

// ShouldIgnoreBranch checks if a branch should be ignored based on a list of patterns.
func ShouldIgnoreBranch(branchName string, patterns []string) bool {
	for _, pattern := range patterns {
		if MatchesBranchPattern(branchName, pattern) {
			return true
		}
	}
	return false
}

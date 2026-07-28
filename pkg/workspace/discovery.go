// Package workspace provides workspace management functionality.
package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DiscoverMainRepos finds all git repositories in the repos directory.
func DiscoverMainRepos(reposDir string) ([]RepositorySpec, error) {
	entries, err := os.ReadDir(reposDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read repos directory: %w", err)
	}

	repos := make([]RepositorySpec, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(reposDir, entry.Name())
		gitPath := filepath.Join(repoPath, ".git")

		info, err := os.Stat(gitPath)
		if err != nil || !info.IsDir() {
			continue
		}

		cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
		output, err := cmd.Output()
		if err != nil {
			// The directory is a git repository but its origin can't be
			// resolved — most often a repo that was created locally and
			// never pushed. Skipping it silently makes it simply absent
			// from the workspace with nothing to debug, so say so.
			fmt.Fprintf(os.Stderr, "workspace: skipping %q: no origin remote (%v)\n", entry.Name(), err)

			continue
		}

		repos = append(repos, RepositorySpec{
			Name: entry.Name(),
			URL:  strings.TrimSpace(string(output)),
		})
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}

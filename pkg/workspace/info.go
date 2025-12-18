package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CreateWorkspaceInfo writes workspace metadata to a .workspace-info file.
func CreateWorkspaceInfo(workspacePath, name, repoURL string, cloned, failed int) error {
	infoPath := filepath.Join(workspacePath, ".workspace-info")
	file, err := os.Create(infoPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, _ = fmt.Fprintf(file, "Workspace: workspace-%s\n", name)
	_, _ = fmt.Fprintf(file, "Created: %s\n", time.Now().Format(time.RFC3339))
	if repoURL != "" {
		_, _ = fmt.Fprintf(file, "Repository: %s\n", repoURL)
	}
	if cloned > 0 {
		_, _ = fmt.Fprintf(file, "Projects cloned: %d\n", cloned)
	}
	if failed > 0 {
		_, _ = fmt.Fprintf(file, "Failed clones: %d\n", failed)
	}

	return nil
}

// ReadWorkspaceInfo reads the creation time from a workspace's info file.
func ReadWorkspaceInfo(workspacePath string) (time.Time, error) {
	infoPath := filepath.Join(workspacePath, ".workspace-info")
	file, err := os.Open(infoPath)
	if err != nil {
		return time.Time{}, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Created: ") {
			return time.Parse(time.RFC3339, strings.TrimPrefix(line, "Created: "))
		}
	}

	return time.Time{}, nil
}

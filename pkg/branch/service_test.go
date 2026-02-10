package branch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jcleira/workspace/pkg/workspace"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	wm := workspace.NewManager("/tmp/workspaces", "/tmp/repos", "/tmp/claude")
	patterns := []string{"main", "master"}

	svc := NewService(wm, patterns)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.workspaceManager != wm {
		t.Error("expected service to use provided workspace manager")
	}
	if len(svc.ignorePatterns) != 2 {
		t.Errorf("expected 2 ignore patterns, got %d", len(svc.ignorePatterns))
	}
}

func TestService_BuildWorkspaceMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, workspacesDir string)
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "empty workspaces returns empty mapping",
			setupFunc: func(t *testing.T, workspacesDir string) {},
			wantLen:   0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			workspacesDir := filepath.Join(tmpDir, "workspaces")
			reposDir := filepath.Join(tmpDir, "repos")
			claudeDir := filepath.Join(tmpDir, ".claude")

			if err := os.MkdirAll(workspacesDir, 0o755); err != nil {
				t.Fatal(err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, workspacesDir)
			}

			wm := workspace.NewManager(workspacesDir, reposDir, claudeDir)
			svc := NewService(wm, nil)

			mapping, err := svc.BuildWorkspaceMapping()

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWorkspaceMapping() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(mapping) != tt.wantLen {
				t.Errorf("BuildWorkspaceMapping() returned %d entries, want %d", len(mapping), tt.wantLen)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, reposDir string)
		wantErr   bool
	}{
		{
			name:      "non-existent repos dir returns empty output",
			setupFunc: nil,
			wantErr:   false,
		},
		{
			name: "empty repos dir returns empty output",
			setupFunc: func(t *testing.T, reposDir string) {
				if err := os.MkdirAll(reposDir, 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			workspacesDir := filepath.Join(tmpDir, "workspaces")
			reposDir := filepath.Join(tmpDir, "repos")
			claudeDir := filepath.Join(tmpDir, ".claude")

			if err := os.MkdirAll(workspacesDir, 0o755); err != nil {
				t.Fatal(err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, reposDir)
			}

			wm := workspace.NewManager(workspacesDir, reposDir, claudeDir)
			svc := NewService(wm, nil)

			output, err := svc.List()

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}

			_ = output // use output to avoid unused variable warning
		})
	}
}

func TestService_PlanCleanup(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesDir := filepath.Join(tmpDir, "workspaces")
	reposDir := filepath.Join(tmpDir, "repos")
	claudeDir := filepath.Join(tmpDir, ".claude")

	if err := os.MkdirAll(workspacesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(reposDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wm := workspace.NewManager(workspacesDir, reposDir, claudeDir)
	svc := NewService(wm, nil)

	plan, err := svc.PlanCleanup()
	if err != nil {
		t.Fatalf("PlanCleanup() error = %v", err)
	}

	if plan.OrphanedBranches == nil {
		t.Error("expected non-nil OrphanedBranches slice")
	}
}

func TestService_ExecuteCleanup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		plan         CleanupPlan
		skipBranches []string
		wantDeleted  int
		wantSkipped  int
	}{
		{
			name: "empty plan returns empty result",
			plan: CleanupPlan{
				OrphanedBranches: []OrphanedBranch{},
			},
			skipBranches: nil,
			wantDeleted:  0,
			wantSkipped:  0,
		},
		{
			name: "skipped branches are recorded",
			plan: CleanupPlan{
				OrphanedBranches: []OrphanedBranch{
					{RepoName: "repo1", BranchName: "branch1"},
				},
			},
			skipBranches: []string{"repo1:branch1"},
			wantDeleted:  0,
			wantSkipped:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			workspacesDir := filepath.Join(tmpDir, "workspaces")
			reposDir := filepath.Join(tmpDir, "repos")
			claudeDir := filepath.Join(tmpDir, ".claude")

			wm := workspace.NewManager(workspacesDir, reposDir, claudeDir)
			svc := NewService(wm, nil)

			result, err := svc.ExecuteCleanup(tt.plan, tt.skipBranches)
			if err != nil {
				t.Fatalf("ExecuteCleanup() error = %v", err)
			}

			if len(result.Deleted) != tt.wantDeleted {
				t.Errorf("Deleted count = %d, want %d", len(result.Deleted), tt.wantDeleted)
			}

			if len(result.Skipped) != tt.wantSkipped {
				t.Errorf("Skipped count = %d, want %d", len(result.Skipped), tt.wantSkipped)
			}
		})
	}
}

func TestService_CheckUnpushed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacesDir := filepath.Join(tmpDir, "workspaces")
	reposDir := filepath.Join(tmpDir, "repos")
	claudeDir := filepath.Join(tmpDir, ".claude")

	wm := workspace.NewManager(workspacesDir, reposDir, claudeDir)
	svc := NewService(wm, nil)

	hasUnpushed, count, err := svc.CheckUnpushed("/nonexistent/repo", "main")

	if err == nil {
		t.Log("CheckUnpushed on non-existent repo may or may not error depending on git behavior")
	}

	if hasUnpushed && count == 0 {
		t.Error("if hasUnpushed is true, count should be > 0")
	}
}

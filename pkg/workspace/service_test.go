package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	manager := NewManager("/tmp/workspaces", "/tmp/repos", "/tmp/claude")
	svc := NewService(manager)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.manager != manager {
		t.Error("expected service to use provided manager")
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     CreateInput
		setupFunc func(t *testing.T, workspacesDir, reposDir string)
		wantErr   bool
		checkFunc func(t *testing.T, output CreateOutput, err error)
	}{
		{
			name: "creates workspace with no repos returns error",
			input: CreateInput{
				Name: "test-workspace",
			},
			setupFunc: func(t *testing.T, workspacesDir, reposDir string) {
				if err := os.MkdirAll(reposDir, 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
			checkFunc: func(t *testing.T, output CreateOutput, err error) {
				if err == nil {
					t.Error("expected error for empty repos directory")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			workspacesDir := filepath.Join(tmpDir, "workspaces")
			reposDir := filepath.Join(tmpDir, "repos")
			claudeDir := filepath.Join(tmpDir, ".claude")

			if tt.setupFunc != nil {
				tt.setupFunc(t, workspacesDir, reposDir)
			}

			manager := NewManager(workspacesDir, reposDir, claudeDir)
			svc := NewService(manager)

			output, err := svc.Create(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, output, err)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     DeleteInput
		setupFunc func(t *testing.T, workspacesDir string)
		wantErr   bool
		checkFunc func(t *testing.T, output DeleteOutput, err error)
	}{
		{
			name: "delete default workspace returns error",
			input: DeleteInput{
				Name:           "default",
				DeleteBranches: true,
			},
			wantErr: true,
			checkFunc: func(t *testing.T, output DeleteOutput, err error) {
				if err == nil {
					t.Error("expected error when deleting default workspace")
				}
			},
		},
		{
			name: "delete non-existent workspace returns error",
			input: DeleteInput{
				Name:           "nonexistent",
				DeleteBranches: true,
			},
			wantErr: true,
			checkFunc: func(t *testing.T, output DeleteOutput, err error) {
				if err == nil {
					t.Error("expected error for non-existent workspace")
				}
			},
		},
		{
			name: "delete existing workspace succeeds",
			input: DeleteInput{
				Name:           "test-delete",
				DeleteBranches: false,
			},
			setupFunc: func(t *testing.T, workspacesDir string) {
				wsPath := filepath.Join(workspacesDir, "workspace-test-delete")
				if err := os.MkdirAll(wsPath, 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output DeleteOutput, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !output.Deleted {
					t.Error("expected Deleted to be true")
				}
			},
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

			manager := NewManager(workspacesDir, reposDir, claudeDir)
			svc := NewService(manager)

			output, err := svc.Delete(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, output, err)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, workspacesDir string)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty workspaces directory returns empty list",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "returns existing workspaces",
			setupFunc: func(t *testing.T, workspacesDir string) {
				if err := os.MkdirAll(filepath.Join(workspacesDir, "workspace-test1"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(workspacesDir, "workspace-test2"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "ignores non-workspace directories",
			setupFunc: func(t *testing.T, workspacesDir string) {
				if err := os.MkdirAll(filepath.Join(workspacesDir, "workspace-valid"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(workspacesDir, "not-a-workspace"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 1,
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

			manager := NewManager(workspacesDir, reposDir, claudeDir)
			svc := NewService(manager)

			infos, err := svc.List()

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(infos) != tt.wantCount {
				t.Errorf("List() returned %d workspaces, want %d", len(infos), tt.wantCount)
			}
		})
	}
}

func TestService_GetPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wsName    string
		setupFunc func(t *testing.T, workspacesDir string)
		wantErr   bool
	}{
		{
			name:    "non-existent workspace returns error",
			wsName:  "nonexistent",
			wantErr: true,
		},
		{
			name:   "existing workspace returns path",
			wsName: "myworkspace",
			setupFunc: func(t *testing.T, workspacesDir string) {
				if err := os.MkdirAll(filepath.Join(workspacesDir, "workspace-myworkspace"), 0o755); err != nil {
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
				tt.setupFunc(t, workspacesDir)
			}

			manager := NewManager(workspacesDir, reposDir, claudeDir)
			svc := NewService(manager)

			path, err := svc.GetPath(tt.wsName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && path == "" {
				t.Error("GetPath() returned empty path")
			}
		})
	}
}

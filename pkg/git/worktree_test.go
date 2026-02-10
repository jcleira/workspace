package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initGitRepo(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = path
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = path
	_ = cmd.Run()

	testFile := filepath.Join(path, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestIsWorktree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, path string)
		want      bool
		wantErr   bool
	}{
		{
			name:      "non-existent path returns false",
			setupFunc: nil,
			want:      false,
			wantErr:   false,
		},
		{
			name: "regular git repo returns false",
			setupFunc: func(t *testing.T, path string) {
				initGitRepo(t, path)
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testPath := filepath.Join(tmpDir, "test-repo")

			if tt.setupFunc != nil {
				tt.setupFunc(t, testPath)
			}

			got, err := IsWorktree(testPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsWorktree() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("IsWorktree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateWorktree(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainRepoPath := filepath.Join(tmpDir, "main-repo")
	initGitRepo(t, mainRepoPath)

	worktreePath := filepath.Join(tmpDir, "worktree")
	branchName := "feature-branch"

	err := CreateWorktree(mainRepoPath, worktreePath, branchName)
	if err != nil {
		t.Fatalf("CreateWorktree() error = %v", err)
	}

	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	isWt, err := IsWorktree(worktreePath)
	if err != nil {
		t.Fatalf("IsWorktree() error = %v", err)
	}
	if !isWt {
		t.Error("created path should be a worktree")
	}
}

func TestCreateWorktree_BranchAlreadyExists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainRepoPath := filepath.Join(tmpDir, "main-repo")
	initGitRepo(t, mainRepoPath)

	cmd := exec.Command("git", "branch", "existing-branch")
	cmd.Dir = mainRepoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	worktreePath := filepath.Join(tmpDir, "worktree")
	err := CreateWorktree(mainRepoPath, worktreePath, "existing-branch")

	if err == nil {
		t.Error("expected error when creating worktree with existing branch name")
	}
}

func TestCheckoutExistingBranch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainRepoPath := filepath.Join(tmpDir, "main-repo")
	initGitRepo(t, mainRepoPath)

	cmd := exec.Command("git", "branch", "existing-branch")
	cmd.Dir = mainRepoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	worktreePath := filepath.Join(tmpDir, "worktree")
	err := CheckoutExistingBranch(mainRepoPath, worktreePath, "existing-branch")
	if err != nil {
		t.Fatalf("CheckoutExistingBranch() error = %v", err)
	}

	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}
}

func TestRemoveWorktree(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainRepoPath := filepath.Join(tmpDir, "main-repo")
	initGitRepo(t, mainRepoPath)

	worktreePath := filepath.Join(tmpDir, "worktree")
	if err := CreateWorktree(mainRepoPath, worktreePath, "test-branch"); err != nil {
		t.Fatalf("CreateWorktree() error = %v", err)
	}

	if err := RemoveWorktree(mainRepoPath, worktreePath); err != nil {
		t.Fatalf("RemoveWorktree() error = %v", err)
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree directory should have been removed")
	}
}

func TestListWorktrees(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainRepoPath := filepath.Join(tmpDir, "main-repo")
	initGitRepo(t, mainRepoPath)

	worktrees, err := ListWorktrees(mainRepoPath)
	if err != nil {
		t.Fatalf("ListWorktrees() error = %v", err)
	}

	if len(worktrees) != 1 {
		t.Errorf("expected 1 worktree (main), got %d", len(worktrees))
	}

	worktreePath := filepath.Join(tmpDir, "worktree")
	if err := CreateWorktree(mainRepoPath, worktreePath, "test-branch"); err != nil {
		t.Fatalf("CreateWorktree() error = %v", err)
	}

	worktrees, err = ListWorktrees(mainRepoPath)
	if err != nil {
		t.Fatalf("ListWorktrees() error = %v", err)
	}

	if len(worktrees) != 2 {
		t.Errorf("expected 2 worktrees, got %d", len(worktrees))
	}
}

func TestBranchExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		branchName string
		setupFunc  func(t *testing.T, repoPath string)
		want       bool
	}{
		{
			name:       "main branch exists",
			branchName: "master",
			setupFunc:  nil,
			want:       true,
		},
		{
			name:       "non-existent branch",
			branchName: "nonexistent-branch",
			setupFunc:  nil,
			want:       false,
		},
		{
			name:       "created branch exists",
			branchName: "feature-branch",
			setupFunc: func(t *testing.T, repoPath string) {
				cmd := exec.Command("git", "branch", "feature-branch")
				cmd.Dir = repoPath
				_ = cmd.Run()
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			repoPath := filepath.Join(tmpDir, "repo")
			initGitRepo(t, repoPath)

			if tt.setupFunc != nil {
				tt.setupFunc(t, repoPath)
			}

			got, err := BranchExists(repoPath, tt.branchName)
			if err != nil {
				t.Fatalf("BranchExists() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("BranchExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBranchCheckedOut(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainRepoPath := filepath.Join(tmpDir, "main-repo")
	initGitRepo(t, mainRepoPath)

	worktreePath := filepath.Join(tmpDir, "worktree")
	branchName := "feature-branch"
	if err := CreateWorktree(mainRepoPath, worktreePath, branchName); err != nil {
		t.Fatalf("CreateWorktree() error = %v", err)
	}

	checkedOut, location, err := IsBranchCheckedOut(mainRepoPath, branchName)
	if err != nil {
		t.Fatalf("IsBranchCheckedOut() error = %v", err)
	}

	if !checkedOut {
		t.Error("expected branch to be checked out")
	}

	realWorktreePath, _ := filepath.EvalSymlinks(worktreePath)
	realLocation, _ := filepath.EvalSymlinks(location)
	if realLocation != realWorktreePath {
		t.Errorf("location = %s, want %s", realLocation, realWorktreePath)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	initGitRepo(t, repoPath)

	branch, err := GetCurrentBranch(repoPath)
	if err != nil {
		t.Fatalf("GetCurrentBranch() error = %v", err)
	}

	if branch != "master" {
		t.Errorf("GetCurrentBranch() = %s, want master", branch)
	}
}

func TestGetAllBranches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	initGitRepo(t, repoPath)

	cmd := exec.Command("git", "branch", "branch1")
	cmd.Dir = repoPath
	_ = cmd.Run()

	cmd = exec.Command("git", "branch", "branch2")
	cmd.Dir = repoPath
	_ = cmd.Run()

	branches, err := GetAllBranches(repoPath)
	if err != nil {
		t.Fatalf("GetAllBranches() error = %v", err)
	}

	if len(branches) != 3 {
		t.Errorf("expected 3 branches, got %d: %v", len(branches), branches)
	}
}

func TestDeleteBranch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	initGitRepo(t, repoPath)

	cmd := exec.Command("git", "branch", "to-delete")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	if err := DeleteBranch(repoPath, "to-delete"); err != nil {
		t.Fatalf("DeleteBranch() error = %v", err)
	}

	exists, err := BranchExists(repoPath, "to-delete")
	if err != nil {
		t.Fatalf("BranchExists() error = %v", err)
	}

	if exists {
		t.Error("branch should have been deleted")
	}
}

func TestDeleteBranch_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	initGitRepo(t, repoPath)

	err := DeleteBranch(repoPath, "nonexistent")
	if err != nil {
		t.Error("DeleteBranch() should not error for non-existent branch")
	}
}

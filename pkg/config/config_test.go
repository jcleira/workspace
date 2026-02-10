package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigManager(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}
	if cm == nil {
		t.Fatal("expected non-nil ConfigManager")
	}
	if cm.config == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestConfigManager_GetConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	cfg := cm.GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if cfg.WorkspacesDir == "" {
		t.Error("expected non-empty WorkspacesDir")
	}
	if cfg.ReposDir == "" {
		t.Error("expected non-empty ReposDir")
	}
	if cfg.ClaudeDir == "" {
		t.Error("expected non-empty ClaudeDir")
	}
}

func TestConfigManager_SetWorkspacesDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	newDir := filepath.Join(tmpDir, "new-workspaces")
	if err := cm.SetWorkspacesDir(newDir); err != nil {
		t.Fatalf("SetWorkspacesDir() error = %v", err)
	}

	cfg := cm.GetConfig()
	if cfg.WorkspacesDir != newDir {
		t.Errorf("WorkspacesDir = %s, want %s", cfg.WorkspacesDir, newDir)
	}
}

func TestConfigManager_IgnoredBranches(t *testing.T) {
	tests := []struct {
		name        string
		operations  func(cm *ConfigManager) error
		wantCount   int
		wantPattern string
	}{
		{
			name: "add ignored branch pattern",
			operations: func(cm *ConfigManager) error {
				return cm.AddIgnoredBranch("feature-*")
			},
			wantCount:   1,
			wantPattern: "feature-*",
		},
		{
			name: "remove ignored branch pattern",
			operations: func(cm *ConfigManager) error {
				if err := cm.AddIgnoredBranch("test-*"); err != nil {
					return err
				}
				return cm.RemoveIgnoredBranch("test-*")
			},
			wantCount: 0,
		},
		{
			name: "clear all patterns",
			operations: func(cm *ConfigManager) error {
				if err := cm.AddIgnoredBranch("pattern1"); err != nil {
					return err
				}
				if err := cm.AddIgnoredBranch("pattern2"); err != nil {
					return err
				}
				return cm.ClearIgnoredBranches()
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configDir := filepath.Join(tmpDir, "config")
			t.Setenv("XDG_CONFIG_HOME", configDir)

			cm, err := NewConfigManager()
			if err != nil {
				t.Fatalf("NewConfigManager() error = %v", err)
			}

			if err := tt.operations(cm); err != nil {
				t.Fatalf("operations error = %v", err)
			}

			patterns := cm.GetIgnoredBranches()
			if len(patterns) != tt.wantCount {
				t.Errorf("got %d patterns, want %d", len(patterns), tt.wantCount)
			}

			if tt.wantPattern != "" && (len(patterns) == 0 || patterns[0] != tt.wantPattern) {
				t.Errorf("pattern = %v, want %s", patterns, tt.wantPattern)
			}
		})
	}
}

func TestConfigManager_AddIgnoredBranch_EmptyPattern(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	if err := cm.AddIgnoredBranch(""); err == nil {
		t.Error("expected error for empty pattern")
	}
}

func TestConfigManager_AddIgnoredBranch_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	if err := cm.AddIgnoredBranch("test-*"); err != nil {
		t.Fatalf("first AddIgnoredBranch() error = %v", err)
	}

	if err := cm.AddIgnoredBranch("test-*"); err == nil {
		t.Error("expected error for duplicate pattern")
	}
}

func TestConfigManager_RemoveIgnoredBranch_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	if err := cm.RemoveIgnoredBranch("nonexistent"); err == nil {
		t.Error("expected error for non-existent pattern")
	}
}

func TestConfigManager_IsInitialized(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(cm *ConfigManager, tmpDir string)
		want      bool
	}{
		{
			name:      "not initialized by default",
			setupFunc: func(cm *ConfigManager, tmpDir string) {},
			want:      false,
		},
		{
			name: "initialized when flag is set",
			setupFunc: func(cm *ConfigManager, tmpDir string) {
				_ = cm.SetInitialized(true)
			},
			want: true,
		},
		{
			name: "initialized when repos dir has content",
			setupFunc: func(cm *ConfigManager, tmpDir string) {
				reposDir := cm.GetConfig().ReposDir
				_ = os.MkdirAll(filepath.Join(reposDir, "some-repo"), 0o755)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configDir := filepath.Join(tmpDir, "config")
			t.Setenv("XDG_CONFIG_HOME", configDir)

			homeDir := filepath.Join(tmpDir, "home")
			t.Setenv("HOME", homeDir)

			cm, err := NewConfigManager()
			if err != nil {
				t.Fatalf("NewConfigManager() error = %v", err)
			}

			tt.setupFunc(cm, tmpDir)

			if got := cm.IsInitialized(); got != tt.want {
				t.Errorf("IsInitialized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigManager_UpdateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cm, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	newWorkspaces := filepath.Join(tmpDir, "ws")
	newRepos := filepath.Join(tmpDir, "repos")
	newClaude := filepath.Join(tmpDir, "claude")

	if err := cm.UpdateConfig(newWorkspaces, newRepos, newClaude); err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}

	cfg := cm.GetConfig()
	if cfg.WorkspacesDir != newWorkspaces {
		t.Errorf("WorkspacesDir = %s, want %s", cfg.WorkspacesDir, newWorkspaces)
	}
	if cfg.ReposDir != newRepos {
		t.Errorf("ReposDir = %s, want %s", cfg.ReposDir, newRepos)
	}
	if cfg.ClaudeDir != newClaude {
		t.Errorf("ClaudeDir = %s, want %s", cfg.ClaudeDir, newClaude)
	}
}

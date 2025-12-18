// Package config provides configuration management for the workspace CLI.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the workspace configuration
type Config struct {
	WorkspacesDir   string   `json:"workspaces_dir"`
	ReposDir        string   `json:"repos_dir"`
	ClaudeDir       string   `json:"claude_dir"`
	IgnoredBranches []string `json:"ignored_branches,omitempty"`
}

// ConfigManager handles configuration operations
type ConfigManager struct {
	configPath string
	config     *Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	workspaceConfigDir := filepath.Join(configDir, "workspace")
	configPath := filepath.Join(workspaceConfigDir, "config.json")

	if err := os.MkdirAll(workspaceConfigDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	cm := &ConfigManager{
		configPath: configPath,
	}

	if err := cm.loadConfig(); err != nil {
		return nil, err
	}

	return cm, nil
}

// loadConfig loads configuration from file or creates default
func (cm *ConfigManager) loadConfig() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		homeDir := os.Getenv("HOME")
		cm.config = &Config{
			WorkspacesDir: filepath.Join(homeDir, "workspaces"),
			ReposDir:      filepath.Join(homeDir, "repos"),
			ClaudeDir:     filepath.Join(homeDir, ".claude"),
		}
		return cm.saveConfig()
	}

	file, err := os.Open(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	cm.config = &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cm.config); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	homeDir := os.Getenv("HOME")
	if cm.config.WorkspacesDir == "" {
		cm.config.WorkspacesDir = filepath.Join(homeDir, "workspaces")
	}
	if cm.config.ReposDir == "" {
		cm.config.ReposDir = filepath.Join(homeDir, "repos")
	}
	if cm.config.ClaudeDir == "" {
		cm.config.ClaudeDir = filepath.Join(homeDir, ".claude")
	}

	return nil
}

// saveConfig saves configuration to file
func (cm *ConfigManager) saveConfig() error {
	file, err := os.Create(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cm.config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetWorkspacesDir sets the workspaces directory
func (cm *ConfigManager) SetWorkspacesDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	cm.config.WorkspacesDir = absDir
	return cm.saveConfig()
}

// SetReposDir sets the repos directory
func (cm *ConfigManager) SetReposDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	cm.config.ReposDir = absDir
	return cm.saveConfig()
}

// SetClaudeDir sets the claude directory
func (cm *ConfigManager) SetClaudeDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	cm.config.ClaudeDir = absDir
	return cm.saveConfig()
}

// GetConfigPath returns the configuration file path
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

// GetIgnoredBranches returns the list of ignored branch patterns
func (cm *ConfigManager) GetIgnoredBranches() []string {
	if cm.config.IgnoredBranches == nil {
		return []string{}
	}
	return cm.config.IgnoredBranches
}

// AddIgnoredBranch adds a branch pattern to the ignore list
func (cm *ConfigManager) AddIgnoredBranch(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	for _, existing := range cm.config.IgnoredBranches {
		if existing == pattern {
			return fmt.Errorf("pattern '%s' already exists", pattern)
		}
	}

	cm.config.IgnoredBranches = append(cm.config.IgnoredBranches, pattern)
	return cm.saveConfig()
}

// RemoveIgnoredBranch removes a branch pattern from the ignore list
func (cm *ConfigManager) RemoveIgnoredBranch(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	newPatterns := make([]string, 0, len(cm.config.IgnoredBranches))
	found := false
	for _, existing := range cm.config.IgnoredBranches {
		if existing == pattern {
			found = true
			continue
		}
		newPatterns = append(newPatterns, existing)
	}

	if !found {
		return fmt.Errorf("pattern '%s' not found", pattern)
	}

	cm.config.IgnoredBranches = newPatterns
	return cm.saveConfig()
}

// SetIgnoredBranches sets the entire list of ignored branches
func (cm *ConfigManager) SetIgnoredBranches(patterns []string) error {
	cm.config.IgnoredBranches = patterns
	return cm.saveConfig()
}

// ClearIgnoredBranches clears all ignored branch patterns
func (cm *ConfigManager) ClearIgnoredBranches() error {
	cm.config.IgnoredBranches = []string{}
	return cm.saveConfig()
}

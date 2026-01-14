# Workspace

<img width="1797" height="1131" alt="workspace" src="https://github.com/user-attachments/assets/11b9a166-7aec-41b0-b23a-1ad4bbe911f5" />

A command-line tool for managing isolated development workspaces with multiple git repositories. Designed for developers working on microservices, monorepos, or multiple related projects.

![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-macOS%20|%20Linux%20|%20Windows-lightgrey)
[![AUR version](https://img.shields.io/aur/version/workspace-cli-bin)](https://aur.archlinux.org/packages/workspace-cli-bin)

## Features

- **Isolated Workspaces**: Create separate workspace environments for different features or experiments
- **Git Worktree Integration**: Feature workspaces use git worktrees for lightweight, isolated branch development
- **Branch Management**: Track, cleanup, and manage branches across all repositories
- **Shared Context**: All workspaces share `.claude` directory for consistent AI assistance
- **Automatic Sync**: Fetches and pulls latest changes when creating workspaces
- **Interactive UI**: Terminal interface powered by Bubble Tea
- **Shell Integration**: Seamless navigation with the `w` command
- **Protected Workspaces**: Default workspace is protected from accidental deletion
- **Unpushed Commit Detection**: Warns before deleting branches with unpushed work

## Installation

### Homebrew (macOS/Linux)

```bash
brew install jcleira/tap/workspace
```

### APT (Debian/Ubuntu)

```bash
curl -LO https://github.com/jcleira/workspace/releases/latest/download/workspace_$(curl -s https://api.github.com/repos/jcleira/workspace/releases/latest | grep tag_name | cut -d '"' -f 4 | tr -d 'v')_linux_amd64.deb
sudo dpkg -i workspace_*.deb
```

### RPM (Fedora/RHEL)

```bash
curl -LO https://github.com/jcleira/workspace/releases/latest/download/workspace_$(curl -s https://api.github.com/repos/jcleira/workspace/releases/latest | grep tag_name | cut -d '"' -f 4 | tr -d 'v')_linux_amd64.rpm
sudo dnf install ./workspace_*.rpm
```

### APK (Alpine)

```bash
curl -LO https://github.com/jcleira/workspace/releases/latest/download/workspace_$(curl -s https://api.github.com/repos/jcleira/workspace/releases/latest | grep tag_name | cut -d '"' -f 4 | tr -d 'v')_linux_amd64.apk
sudo apk add --allow-untrusted workspace_*.apk
```

### Scoop (Windows)

```powershell
scoop bucket add jcleira https://github.com/jcleira/scoop-bucket
scoop install workspace
```

### AUR (Arch Linux)

```bash
yay -S workspace-cli-bin
# or
paru -S workspace-cli-bin
```

### Go Install

```bash
go install github.com/jcleira/workspace@latest
```

### From Source

```bash
git clone https://github.com/jcleira/workspace.git
cd workspace
make install

# Or install to user directory
make install-local
```

## Quick Start

### 1. Initial Setup

```bash
# Configure workspace directories
workspace config setup

# View current configuration
workspace config show
```

### 2. Create Your First Workspace

```bash
# Create a new workspace (uses git worktrees)
workspace create my-feature

```

### 3. Set Up Shell Integration

```bash
# Initialize shell integration for the 'w' command
workspace config init

# Add to your shell config and reload
source ~/.bashrc  # or ~/.zshrc
```

### 4. Navigate Between Workspaces

```bash
# Interactive workspace selector
workspace  # or just 'w' after shell integration

# Direct navigation
w my-feature

# List all workspaces
workspace list
```

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `workspace` | Interactive workspace selector |
| `workspace create <name>` | Create a new workspace |
| `workspace list` | List all workspaces |
| `workspace switch <name>` | Switch to a workspace |
| `workspace delete <name>` | Delete workspace and its branches |

### Branch Management Commands

| Command | Description |
|---------|-------------|
| `workspace branch list` | List all branches with workspace associations |
| `workspace branch cleanup` | Delete orphaned branches (interactive) |
| `workspace branch ignore add <pattern>` | Add pattern to ignore list |
| `workspace branch ignore remove <pattern>` | Remove pattern from ignore list |
| `workspace branch ignore list` | List ignored patterns |
| `workspace branch ignore clear` | Clear all ignored patterns |

### Configuration Commands

| Command | Description |
|---------|-------------|
| `workspace config show` | Display current configuration |
| `workspace config setup` | Interactive configuration setup |
| `workspace config set <key> <value>` | Set configuration value |
| `workspace config init` | Initialize shell integration |
| `workspace config completion <shell>` | Generate shell completions |

### Configuration Keys

- `workspaces-dir`: Directory where workspaces are stored
- `repos-dir`: Directory where main git repositories are located
- `claude-dir`: Shared Claude context directory

## Git Integration

### Automatic Remote Sync

When creating a new workspace, the CLI automatically:
1. Fetches latest changes from remote (`git fetch origin`)
2. Pulls the default branch to ensure it's up-to-date
3. Creates worktrees from the latest code

### Git Worktree vs Normal Clones

- **default workspace**: Uses normal git clones for persistent, long-term development
- **feature workspaces**: Uses git worktrees for lightweight branch checkouts

```bash
# Creates normal clones
workspace create default

# Creates worktrees linked to repos-dir
workspace create feature-auth
```

### Branch Lifecycle

**On Creation:**
- Branches are created from the latest main/master

**On Deletion:**
- Branches are deleted by default when deleting workspace
- Branches with unpushed commits are automatically skipped (safe default)

### Branch Ignore Patterns

Protect branches from cleanup with patterns:

```bash
# Add patterns
workspace branch ignore add "runtime-*"
workspace branch ignore add "staging"

# Pattern types supported
workspace branch ignore add "local-*"     # Prefix
workspace branch ignore add "*-backup"    # Suffix
workspace branch ignore add "release/*"   # Glob
```

## Project Structure

```
workspace/
├── main.go                 # Entry point
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command & dashboard launcher
│   ├── create/            # Workspace creation
│   ├── delete/            # Workspace deletion
│   ├── list/              # List workspaces
│   ├── switch/            # Switch workspace
│   ├── branch/            # Branch management (list, cleanup, ignore)
│   └── config/            # Configuration (show, set, setup, init)
├── pkg/                    # Public packages
│   ├── workspace/         # Workspace service, manager, types
│   ├── branch/            # Branch service and types
│   ├── status/            # Repository status service
│   ├── config/            # Configuration management
│   ├── git/               # Git operations (clone, worktree, status)
│   ├── ui/
│   │   ├── commands/      # CLI output helpers
│   │   └── dashboard/     # Interactive dashboard (Bubble Tea)
│   └── shell/             # Shell integration
├── Makefile               # Build automation
└── README.md              # This file
```

## Development

### Building from Source

```bash
git clone https://github.com/jcleira/workspace.git
cd workspace
go mod download
make build
make test
```

### Makefile Targets

```bash
make build         # Build the binary
make build-all     # Build for all platforms
make install       # Install to /usr/local/bin
make install-local # Install to ~/go/bin
make test          # Run tests
make test-coverage # Generate coverage report
make check         # Run fmt, vet, and lint
make lint          # Run golangci-lint
make clean         # Remove build artifacts
```

## Configuration

Configuration is stored in `~/.config/workspace/config.json`:

```json
{
  "workspaces_dir": "/Users/you/workspaces",
  "repos_dir": "/Users/you/repos",
  "claude_dir": "/Users/you/.claude"
}
```

### Manual Configuration

```bash
workspace config set workspaces-dir ~/Projects/workspaces
workspace config set repos-dir ~/Projects/repos
workspace config set claude-dir ~/Projects/.claude
```

## Shell Integration

The `w` function provides quick workspace navigation:

```bash
w              # Interactive workspace selection
w project      # Navigate directly to workspace
w pr[TAB]      # Autocomplete workspace names
```

### Shell Completion

```bash
# Bash
workspace config completion bash > /etc/bash_completion.d/workspace

# Zsh
workspace config completion zsh > "${fpath[1]}/_workspace"

# Fish
workspace config completion fish > ~/.config/fish/completions/workspace.fish
```

## Usage Examples

### Creating Workspaces

```bash
# Create a feature workspace
workspace create frontend

# Default workspace uses normal clones
workspace create default
```

### Managing Branches

```bash
# List all branches and their workspaces
workspace branch list

# Clean up orphaned branches
workspace branch cleanup

# Protect branches from cleanup
workspace branch ignore add "staging"
workspace branch ignore add "release-*"
```

### Deleting Workspaces

```bash
# Delete workspace and its branches
workspace delete old-feature
```

## Claude Integration

Workspaces automatically link to a shared `.claude` directory:

```
workspace-myproject/
├── .claude -> ~/Tactic/.claude  # Shared AI context
├── frontend/                    # Your repositories
├── backend/
└── .workspace-info              # Workspace metadata
```

## Troubleshooting

### Command not found

```bash
export PATH="/usr/local/bin:$PATH"
# Or for user installation
export PATH="$HOME/go/bin:$PATH"
```

### Shell integration not working

1. Run `workspace config init`
2. Add the function to your shell config
3. Reload: `source ~/.bashrc` or `source ~/.zshrc`

### Git authentication issues

```bash
ssh-add ~/.ssh/id_rsa
ssh -T git@github.com
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

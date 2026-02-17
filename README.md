# Workspace

<img width="1797" height="1131" alt="workspace" src="https://github.com/user-attachments/assets/11b9a166-7aec-41b0-b23a-1ad4bbe911f5" />

Workspace CLI is a Go-based tool that keeps microservice and monorepo work tidy by carving out isolated workspaces with synchronized git worktrees, shared AI context, and a fast terminal UX.

![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-macOS%20|%20Linux%20|%20Windows-lightgrey)
[![AUR version](https://img.shields.io/aur/version/workspace-cli-bin)](https://aur.archlinux.org/packages/workspace-cli-bin)

## Why Workspace CLI

- Keep parallel feature streams isolated while sharing a single repo checkout via git worktrees
- Guarantee every workspace starts from a freshly synced main branch
- Jump between workspaces (and shells) with the built-in `w` helper
- Track orphaned or stale branches across all repos and clean them up safely
- Share `.claude` data across every workspace for consistent AI context

## Key Capabilities

- **Workspace automation**: Create, list, switch, and delete feature workspaces in seconds
- **Git awareness**: Auto-fetch, detect unpushed commits, and manage worktrees without manual scripting
- **Branch hygiene**: Interactive cleanup plus pattern-based ignore rules for protected branches
- **Terminal UX**: Bubble Tea dashboard for glanceable status plus rich CLI output helpers
- **Shell integration**: `workspace config init` wires the `w` function and autocompletion into bash/zsh/fish

## How It Works

- `workspaces-dir`: root that holds each `workspace-<name>` directory
- `repos-dir`: canonical clones that power git worktrees for every feature workspace
- `claude-dir`: shared AI state, symlinked into every workspace as `.claude`
- `workspace create <name>` syncs origin/main, creates worktrees, and links metadata so delete/cleanup know what to remove later
- The protected `workspace-default` uses full clones (no worktrees) for long-lived work

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
# or
paru -S workspace-cli-bin
```

### Go Install

```bash
```

### From Source

```bash
cd workspace
make install

# Or install to user directory
make install-local
```

## Quick Start

```bash
# 1. Configure directories (workspaces, repos, claude)
workspace config setup

# 2. Create a feature workspace (git worktree)
workspace create my-feature

# 3. Enable shell helpers
workspace config init && source ~/.bashrc   # or ~/.zshrc / ~/.config/fish/config.fish

# 4. Jump between workspaces
w             # interactive selector
w my-feature  # jump directly
workspace list
```

## Command Reference

### Core

| Command | Description |
|---------|-------------|
| `workspace` | Bubble Tea dashboard / selector |
| `workspace create <name>` | Create workspace and git worktrees |
| `workspace list` | Show all workspaces |
| `workspace switch <name>` | Change into workspace directory |
| `workspace delete <name>` | Remove workspace and branches |

### Branch Management

| Command | Description |
|---------|-------------|
| `workspace branch list` | Show branches + owning workspaces, unpushed counts |
| `workspace branch cleanup` | Interactive orphan cleanup |
| `workspace branch ignore add/remove/list/clear` | Manage protected patterns |

### Configuration

| Command | Description |
|---------|-------------|
| `workspace config show` | Print resolved directories |
| `workspace config setup` | Guided setup wizard |
| `workspace config set <key> <value>` | Update a single key |
| `workspace config init` | Install `w` helper + completions |
| `workspace config completion <shell>` | Emit shell-specific completion script |

Configuration lives in `~/.config/workspace/config.json`:

```json
{
  "workspaces_dir": "/Users/you/workspaces",
  "repos_dir": "/Users/you/repos",
  "claude_dir": "/Users/you/.claude"
}
```

## Workspace Lifecycle

- **Create**: `workspace create feature-auth` fetches origin, updates main, makes worktrees, and records metadata
- **Work**: `w feature-auth` drops you into the workspace where every repo is checked out on the new branch
- **Manage branches**: `workspace branch list` surfaces stale/orphaned branches with unpushed counts and last commit info
- **Cleanup**: `workspace delete feature-auth` removes the workspace and (optionally) the branches unless they have unpushed work
- **Protect**: `workspace branch ignore add "runtime-*"` keeps long-lived branches off cleanup lists

## Shared Claude Context

```
workspace-myproject/
├── .claude -> ~/Tactic/.claude
├── frontend/
├── backend/
└── .workspace-info
```

Every workspace symlinks `.claude`, so AI tools keep the same context even when jumping between branches.

## Project Layout

```
workspace/
├── main.go
├── cmd/
│   ├── root.go
│   ├── create/
│   ├── delete/
│   ├── list/
│   ├── switch/
│   ├── branch/
│   └── config/
├── pkg/
│   ├── workspace/
│   ├── branch/
│   ├── status/
│   ├── config/
│   ├── git/
│   ├── ui/
│   │   ├── commands/
│   │   └── dashboard/
│   └── shell/
├── Makefile
└── README.md
```

## Development

```bash
git clone https://github.com/jcleira/workspace.git
cd workspace
go mod download
make build      # compile workspace CLI
make test       # run unit tests with race detector
make check      # fmt + vet + golangci-lint
```

Additional targets: `make build-all`, `make install`, `make install-local`, `make test-coverage`, `make clean`.

## Troubleshooting

- **Command not found**: ensure `/usr/local/bin` or `$HOME/go/bin` is on `PATH`
- **Shell integration missing**: rerun `workspace config init`, append output to shell rc, reload shell
- **Git auth failures**: verify SSH keys via `ssh-add ~/.ssh/id_rsa` and `ssh -T git@github.com`

## License

MIT License – see [LICENSE](LICENSE).

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Lipgloss](https://github.com/charmbracelet/lipgloss)

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Workspace CLI is a Go-based command-line tool for managing isolated development workspaces with multiple git repositories. It's designed for developers working on microservices, monorepos, or multiple related projects, providing seamless navigation and project management.

## Architecture

The codebase follows clean architecture principles with clear separation of concerns:

- **cmd/**: Cobra commands for CLI functionality
  - `root.go`: Root command & interactive selector
  - `create/`: Workspace creation with auto-sync
  - `list/`: List workspaces
  - `switch/`: Switch workspace
  - `delete/`: Delete workspace with automatic branch cleanup
  - `config/`: Configuration management (show, set, setup, init, completion)
  - `branch/`: Branch management (list, cleanup, ignore)
- **pkg/**: Public packages with core business logic
  - **workspace/**: Core workspace operations (Service, Manager, types)
  - **branch/**: Branch management service and types
  - **status/**: Repository status service and types
  - **config/**: Configuration file management
  - **git/**: Git repository operations (clone, worktree, status, patterns)
  - **ui/**: Terminal UI components
    - **commands/**: CLI output helpers (PrintSuccess, PrintError, prompts)
    - **dashboard/**: Interactive dashboard UI (Bubble Tea)
  - **shell/**: Shell integration and navigation
- **main.go**: Minimal entry point

## Essential Commands

### Building and Testing
```bash
make build          # Build the binary
make build-all      # Build for all platforms (Linux/macOS, AMD64/ARM64)
make test           # Run all tests with race detection
make test-coverage  # Generate coverage report (coverage.html)
make check          # Run fmt, vet, and lint
make fmt            # Format code with gofmt
make vet            # Run go vet
make lint           # Run golangci-lint
make lint-fix       # Auto-fix linting issues
make lint-install   # Install golangci-lint
```

### Installation
```bash
make install        # Install to /usr/local/bin (requires sudo)
make install-local  # Install to ~/go/bin
make uninstall      # Remove from /usr/local/bin and config
make clean          # Remove build artifacts and dist/
```

### Development
```bash
go run main.go                         # Run interactive selector
go run main.go list                    # Test list command
go run main.go create test-workspace   # Test create command
go run main.go config show             # View configuration
go test ./pkg/... -v              # Run specific package tests
go test -race ./...                    # Run all tests with race detection
```

### Dependencies
```bash
make deps           # Download and tidy dependencies
make update         # Update all dependencies to latest versions
```

## Code Style Guidelines

**CRITICAL**: This project follows a strict no-comments-in-functions policy:
- DO NOT add comments inside function bodies
- Code should be self-documenting through clear naming
- Only use documentation comments above function signatures
- Use descriptive variable and function names

## Adding New Features

### Adding a New Command

1. Create a new directory in `cmd/` (e.g., `cmd/mycommand/`)
2. Create the main command file (e.g., `cmd/mycommand/mycommand.go`)
3. Define the cobra command structure with an exported `Cmd` variable
4. Register it in `cmd/root.go` init() function
5. Use existing packages from `pkg/` for functionality

Example structure:
```go
package mycommand

import (
    "github.com/spf13/cobra"
    "github.com/jcleira/workspace/pkg/ui/commands"
    "github.com/jcleira/workspace/pkg/workspace"
)

var Cmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation using commands.PrintSuccess(), etc.
    },
}
```

Then in `cmd/root.go`:
```go
import "github.com/jcleira/workspace/cmd/mycommand"

func init() {
    RootCmd.AddCommand(mycommand.Cmd)
}
```

### Core Components

- **Service** (`pkg/workspace/service.go`): High-level workspace operations (Create, Delete, List)
- **Manager** (`pkg/workspace/manager.go`): Low-level workspace directory operations
- **BranchService** (`pkg/branch/service.go`): Branch listing, cleanup, and ignore patterns
- **ConfigManager** (`pkg/config/config.go`): Manages ~/.config/workspace/config.json
- **Commands Package** (`pkg/ui/commands/`): CLI output (PrintSuccess, PrintError, PrintInfo, PromptYesNo)
- **Dashboard** (`pkg/ui/dashboard/`): Bubble Tea-based interactive workspace dashboard
- **Shell Functions** (`pkg/shell/functions.go`): Generates shell integration scripts

## Testing Guidelines

- Write table-driven tests for functions with multiple cases
- Use the standard `testing` package (no external test frameworks)
- Mock external dependencies (git, filesystem) when needed
- Test files should be in the same package
- Run `make test` before committing
- Check coverage with `make test-coverage`

## Error Handling

- Always check and handle errors explicitly
- Use `fmt.Errorf` with `%w` for error wrapping
- Display user-friendly error messages via `commands.PrintError()`
- Return errors up the stack, handle at command level
- Error messages: lowercase, no trailing punctuation

## Key Dependencies

- **cobra** v1.9.1: Command-line interface framework
- **bubbletea** v1.3.6: Terminal UI framework for interactive mode
- **lipgloss** v1.1.0: Terminal styling
- **bubbles** v0.21.0: UI components for Bubble Tea
- **No database**: All data stored in filesystem

## Common Development Tasks

### Running a Single Test
```bash
go test -v -run TestFunctionName ./pkg/workspace/
```

### Debugging
```bash
go run -race main.go list           # Check for race conditions
go build -gcflags="-m" .            # Check compiler optimizations
```

### Workspace Operations
```bash
# Workspace Management
workspace create myfeature                    # Create workspace (auto-syncs main repos)
workspace list                                # List all workspaces
workspace switch myfeature                    # Switch to workspace
workspace delete oldfeature                   # Delete workspace and its branches

# Branch Management
workspace branch list                         # List all branches and their workspaces
workspace branch cleanup                      # Delete orphaned branches (interactive)
```

## Git Integration and Branch Management

### Automatic Remote Sync
When creating a new workspace (except "default"), the CLI automatically:
1. Fetches latest changes from remote (`git fetch origin`)
2. Pulls the default branch (main/master) to ensure it's up-to-date
3. Shows progress for each repository being synced
4. This ensures new branches are always created from the latest code

### Branch Lifecycle Management

#### On Creation
- New branches are automatically created from the latest main/master
- Detects and prevents creating worktrees for branches already checked out elsewhere

#### On Deletion
- **Default behavior**: Deletes associated git branches when deleting workspace
- Branches with unpushed commits are automatically skipped (safe default)
- Skipped branches are reported to the user with unpushed commit count

### Branch Management Commands

#### `workspace branch list`
- Lists all branches across all main repositories
- Shows which workspace each branch belongs to
- Highlights orphaned branches (no associated workspace)
- Marks ignored branches separately (based on configured patterns)
- Displays unpushed commit count for each branch
- Shows last commit time and author
- Color-coded output for easy scanning

#### `workspace branch cleanup`
- Finds all orphaned branches (branches without workspaces)
- **Automatically skips ignored branches** based on patterns
- Shows preview of branches to be deleted
- Checks for unpushed commits before deletion
- Interactive confirmation for safety
- Can skip individual branches during cleanup

#### `workspace branch ignore` (Pattern-based Branch Protection)
Manage patterns for branches that should never be flagged as orphaned or deleted:

```bash
# Add patterns to ignore
workspace branch ignore add "runtime-*"     # Ignore all branches starting with "runtime-"
workspace branch ignore add "local-*"       # Ignore all branches starting with "local-"
workspace branch ignore add "staging"       # Ignore specific branch name

# List current patterns
workspace branch ignore list

# Remove a pattern
workspace branch ignore remove "runtime-*"

# Clear all patterns
workspace branch ignore clear
```

**Pattern Matching Support:**
- Exact matches: `staging`, `production`
- Prefix patterns: `local-*`, `runtime-*`
- Suffix patterns: `*-backup`, `*-old`
- Glob patterns: `release/*`, `hotfix/*`

**Use Cases:**
- Keep local runtime branches that are used for specific environments
- Protect production/staging branches from accidental cleanup
- Maintain long-lived feature branches without workspace association

## Project Conventions

- **Workspace naming**: Always prefix with "workspace-" internally (e.g., "workspace-frontend")
- **Protected workspace**: "workspace-default" cannot be deleted
- **Config location**: `~/.config/workspace/config.json`
- **Default directories**:
  - Workspaces: `~/Tactic/workspaces/`
  - Main repositories: `~/Tactic/repos/` (for worktree-based workspaces)
  - Shared Claude: `~/Tactic/.claude/`
- **File operations**: Always use absolute paths internally
- **Symlinks**: Each workspace has `.claude -> ~/Tactic/.claude` symlink
- **Branch deletion**: Branches are always deleted when deleting workspaces

## Important Implementation Details

1. **Global state**: ConfigManager and WorkspaceManager are package-level vars in `cmd/` initialized by `InitializeConfig()`
2. **Error checking**: All errors are explicitly checked (enforced by linter)
3. **File closing**: Use `defer func() { _ = file.Close() }()` pattern to handle linter warnings
4. **Shell integration**: Generated functions support bash, zsh, and fish via `workspace config init`
5. **Interactive mode**: Root command without args launches the Bubble Tea dashboard

## Performance Considerations

- Workspace discovery walks directories recursively - cache when possible
- Interactive UI uses Bubble Tea for smooth terminal performance
- Git operations can be slow - show progress to user via ui.PrintInfo()
- Build with `go build -ldflags` to embed version info

## Security Notes

- Never commit sensitive data (keys, tokens)
- User-provided paths are used directly (by design for flexibility)
- Git operations use user's SSH/HTTPS credentials
- Config file has 0644 permissions, workspace dirs have 0755
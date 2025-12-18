package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jcleira/workspace/pkg/ui/commands"
	"github.com/jcleira/workspace/pkg/workspace"
)

// NavigateToWorkspace changes to the workspace directory and starts a new shell.
func NavigateToWorkspace(ws workspace.Workspace) {
	commands.PrintSuccess(fmt.Sprintf("Navigating to: %s", filepath.Base(ws.Path)))

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	if err := os.Chdir(ws.Path); err != nil {
		commands.PrintError(fmt.Sprintf("Failed to change directory: %v", err))
		fmt.Printf("cd %s\n", ws.Path)
		return
	}

	if pwd, err := os.Getwd(); err == nil {
		commands.PrintInfo(fmt.Sprintf("Now in: %s", pwd))
	}

	fmt.Println()
	commands.PrintInfo("Starting new shell in workspace. Type 'exit' to return.")

	cmd := exec.Command(shell)
	cmd.Dir = ws.Path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("WORKSPACE_NAME=%s", filepath.Base(ws.Path)),
		fmt.Sprintf("WORKSPACE_PATH=%s", ws.Path),
	)

	if err := cmd.Run(); err != nil {
		commands.PrintError(fmt.Sprintf("Failed to start shell: %v", err))
	}
}

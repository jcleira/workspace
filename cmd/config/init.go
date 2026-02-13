package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/pkg/shell"
	"github.com/jcleira/workspace/pkg/ui/commands"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize shell integration",
	Long:  `Generate shell functions for seamless workspace navigation.`,
	Run: func(c *cobra.Command, args []string) {
		initShellIntegration()
	},
}

func init() {
	ConfigCmd.AddCommand(initCmd)
}

func initShellIntegration() {
	shellPath := os.Getenv("SHELL")
	shellName := filepath.Base(shellPath)

	homeDir := os.Getenv("HOME")

	var rcFile string
	var functionContent string

	switch shellName {
	case "bash":
		rcFile = filepath.Join(homeDir, ".bashrc")
		functionContent = shell.GenerateBashFunction()
	case "zsh":
		rcFile = filepath.Join(homeDir, ".zshrc")
		functionContent = shell.GenerateZshFunction()
	case "fish":
		fishConfigDir := filepath.Join(homeDir, ".config", "fish", "functions")
		rcFile = filepath.Join(fishConfigDir, "w.fish")
		functionContent = shell.GenerateFishFunction()

		if err := os.MkdirAll(fishConfigDir, 0o755); err != nil {
			commands.PrintErrorf("Failed to create fish config directory: %v", err)
			return
		}
	default:
		commands.PrintErrorf("Unsupported shell: %s", shellName)
		fmt.Println("Supported shells: bash, zsh, fish")
		return
	}

	fmt.Println()
	fmt.Println(commands.TitleStyle.Render("Shell Integration"))
	fmt.Println()

	if shellName == "fish" {
		if err := os.WriteFile(rcFile, []byte(functionContent), 0o644); err != nil {
			commands.PrintErrorf("Failed to write fish function: %v", err)
			return
		}
		commands.PrintSuccessf("Fish function written to: %s", rcFile)
		commands.PrintInfo("Restart your shell or run 'source ~/.config/fish/config.fish' to use the 'w' command")
	} else {
		fmt.Printf("Add this function to your %s:\n\n", commands.InfoStyle.Render(rcFile))
		fmt.Println(functionContent)
		fmt.Println()
		commands.PrintInfof("After adding, restart your shell or run 'source %s' to use the 'w' command", rcFile)
	}

	fmt.Println()
	fmt.Printf("Usage: %s\n", commands.SuccessStyle.Render("w"))
	fmt.Printf("       %s\n", commands.SuccessStyle.Render("w <workspace-name>"))
}

package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current workspace configuration with health check information.`,
	Run: func(_ *cobra.Command, _ []string) {
		showConfig()
	},
}

func init() {
	ConfigCmd.AddCommand(showCmd)
}

func showConfig() {
	config := cmd.ConfigManager.GetConfig()

	fmt.Println()
	fmt.Println(commands.TitleStyle.Render("Workspace Configuration"))
	fmt.Println()

	fmt.Printf("Config file: %s\n", commands.InfoStyle.Render(cmd.ConfigManager.GetConfigPath()))
	fmt.Println()

	workspacesStatus := checkDirStatus(config.WorkspacesDir, "workspaces")
	reposStatus := checkDirStatus(config.ReposDir, "repos")
	claudeStatus := checkDirStatus(config.ClaudeDir, "")

	fmt.Printf("Workspaces directory: %s %s\n", commands.SuccessStyle.Render(config.WorkspacesDir), workspacesStatus)
	fmt.Printf("Repos directory:      %s %s\n", commands.SuccessStyle.Render(config.ReposDir), reposStatus)
	fmt.Printf("Claude directory:     %s %s\n", commands.SuccessStyle.Render(config.ClaudeDir), claudeStatus)
	fmt.Println()
}

func checkDirStatus(path, dirType string) string {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return commands.ColorWarning("(missing)")
	}
	if err != nil {
		return commands.ColorWarning("(error)")
	}
	if !info.IsDir() {
		return commands.ColorWarning("(not a directory)")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return commands.ColorSuccess("(exists)")
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() && e.Name() != ".claude" && !isHiddenDir(e.Name()) {
			count++
		}
	}

	switch dirType {
	case "workspaces":
		return commands.ColorSuccess(fmt.Sprintf("(exists, %d workspace(s))", count))
	case "repos":
		return commands.ColorSuccess(fmt.Sprintf("(exists, %d repo(s))", count))
	default:
		return commands.ColorSuccess("(exists)")
	}
}

func isHiddenDir(name string) bool {
	return name != "" && name[0] == '.'
}

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current workspace configuration.`,
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

	fmt.Printf("Workspaces directory: %s\n", commands.SuccessStyle.Render(config.WorkspacesDir))
	fmt.Printf("Repos directory:      %s\n", commands.SuccessStyle.Render(config.ReposDir))
	fmt.Printf("Claude directory:     %s\n", commands.SuccessStyle.Render(config.ClaudeDir))
	fmt.Println()
}

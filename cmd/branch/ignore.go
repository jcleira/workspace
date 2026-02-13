package branch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/ui/commands"
)

var ignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Manage ignored branch patterns",
	Long:  `Manage patterns for branches that should be ignored in orphan detection and cleanup.`,
}

var ignoreAddCmd = &cobra.Command{
	Use:   "add <pattern>",
	Short: "Add a branch pattern to ignore list",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		pattern := args[0]
		if err := cmd.ConfigManager.AddIgnoredBranch(pattern); err != nil {
			commands.PrintErrorf("Failed to add pattern: %v", err)
			return
		}
		commands.PrintSuccessf("Added pattern '%s' to ignore list", pattern)
	},
}

var ignoreRemoveCmd = &cobra.Command{
	Use:   "remove <pattern>",
	Short: "Remove a branch pattern from ignore list",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		pattern := args[0]
		if err := cmd.ConfigManager.RemoveIgnoredBranch(pattern); err != nil {
			commands.PrintErrorf("Failed to remove pattern: %v", err)
			return
		}
		commands.PrintSuccessf("Removed pattern '%s' from ignore list", pattern)
	},
}

var ignoreListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ignored branch patterns",
	Run: func(_ *cobra.Command, _ []string) {
		patterns := cmd.ConfigManager.GetIgnoredBranches()
		if len(patterns) == 0 {
			commands.PrintInfo("No ignored branch patterns configured")
			return
		}

		commands.PrintInfo("Ignored branch patterns:")
		for _, pattern := range patterns {
			fmt.Printf("  - %s\n", pattern)
		}
	},
}

var ignoreClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all ignored branch patterns",
	Run: func(_ *cobra.Command, _ []string) {
		if err := cmd.ConfigManager.ClearIgnoredBranches(); err != nil {
			commands.PrintErrorf("Failed to clear ignored patterns: %v", err)
			return
		}
		commands.PrintSuccess("Cleared all ignored branch patterns")
	},
}

func init() {
	BranchCmd.AddCommand(ignoreCmd)

	ignoreCmd.AddCommand(ignoreAddCmd)
	ignoreCmd.AddCommand(ignoreRemoveCmd)
	ignoreCmd.AddCommand(ignoreListCmd)
	ignoreCmd.AddCommand(ignoreClearCmd)
}

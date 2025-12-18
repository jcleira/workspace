package branch

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/branch"
	"github.com/jcleira/workspace/pkg/ui/commands"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all branches and their associated workspaces",
	Long:  `List all branches in main repositories, showing which workspaces they belong to and if they have unpushed commits.`,
	Run: func(_ *cobra.Command, _ []string) {
		listBranches()
	},
}

func init() {
	BranchCmd.AddCommand(listCmd)
}

func listBranches() {
	svc := branch.NewService(cmd.WorkspaceManager, cmd.ConfigManager.GetIgnoredBranches())

	output, err := svc.List()
	if err != nil {
		commands.PrintError(fmt.Sprintf("Failed to list branches: %v", err))
		return
	}

	if len(output.Repositories) == 0 {
		commands.PrintInfo("No main repositories found")
		return
	}

	displayBranchList(output)
}

func displayBranchList(output *branch.ListOutput) {
	for _, repo := range output.Repositories {
		fmt.Printf("\n%s\n", commands.ColorInfo(fmt.Sprintf("=== Repository: %s ===", repo.RepoName)))

		for _, b := range repo.Branches {
			if b.IsDefault {
				fmt.Printf("  %s (default branch)\n", commands.ColorSuccess(b.Name))
				continue
			}

			var statusParts []string
			switch {
			case b.WorkspaceName != "":
				statusParts = append(statusParts, fmt.Sprintf("workspace: %s", commands.ColorSuccess(b.WorkspaceName)))
			case b.IsIgnored:
				statusParts = append(statusParts, commands.ColorInfo("ignored"))
			default:
				statusParts = append(statusParts, commands.ColorWarning("orphaned"))
			}

			if b.HasUnpushed {
				statusParts = append(statusParts, commands.ColorWarning(fmt.Sprintf("%d unpushed", b.UnpushedCount)))
			}

			if b.LastCommitTime != "" {
				statusParts = append(statusParts, fmt.Sprintf("last: %s by %s", b.LastCommitTime, b.LastCommitBy))
			}

			fmt.Printf("  %s  [%s]\n", b.Name, strings.Join(statusParts, " | "))
		}
	}

	summary := fmt.Sprintf("Total branches: %d", output.TotalBranches)
	if output.OrphanedCount > 0 {
		summary += fmt.Sprintf(" (orphaned: %d)", output.OrphanedCount)
	}

	if output.IgnoredCount > 0 {
		summary += fmt.Sprintf(" (ignored: %d)", output.IgnoredCount)
	}

	fmt.Printf("\n%s\n", commands.ColorInfo(summary))
}

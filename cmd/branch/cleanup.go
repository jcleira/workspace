package branch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
	"github.com/jcleira/workspace/pkg/branch"
	"github.com/jcleira/workspace/pkg/ui/commands"
)

var cleanupCmd = &cobra.Command{
	Use:     "cleanup",
	Short:   "Delete orphaned branches",
	Long:    `Delete branches that don't have an associated workspace (orphaned branches).`,
	Example: `  workspace branch cleanup`,
	Run: func(_ *cobra.Command, _ []string) {
		cleanupBranches()
	},
}

func init() {
	BranchCmd.AddCommand(cleanupCmd)
}

func cleanupBranches() {
	svc := branch.NewService(cmd.WorkspaceManager, cmd.ConfigManager.GetIgnoredBranches())

	plan, err := svc.PlanCleanup()
	if err != nil {
		commands.PrintErrorf("Failed to plan cleanup: %v", err)
		return
	}

	if len(plan.OrphanedBranches) == 0 {
		if plan.SkippedIgnored > 0 {
			commands.PrintSuccessf("No orphaned branches found! (%d ignored branches skipped)", plan.SkippedIgnored)
		} else {
			commands.PrintSuccess("No orphaned branches found!")
		}
		return
	}

	commands.PrintWarningf("Found %d orphaned branch(es):", len(plan.OrphanedBranches))
	if plan.SkippedIgnored > 0 {
		commands.PrintInfof("(%d ignored branches skipped)", plan.SkippedIgnored)
	}
	for _, ob := range plan.OrphanedBranches {
		if ob.HasUnpushed {
			fmt.Printf("  - %s: %s %s\n", ob.RepoName, ob.BranchName, commands.ColorWarning(fmt.Sprintf("(%d unpushed)", ob.UnpushedCount)))
		} else {
			fmt.Printf("  - %s: %s\n", ob.RepoName, ob.BranchName)
		}
	}

	if !commands.PromptYesNo("\nDo you want to delete these branches? (y/n): ") {
		commands.PrintInfo("Cleanup canceled")
		return
	}

	skipBranches := promptForUnpushedBranches(plan)

	result, err := svc.ExecuteCleanup(plan, skipBranches)
	if err != nil {
		commands.PrintErrorf("Failed to execute cleanup: %v", err)
		return
	}

	displayCleanupResult(result)
}

func promptForUnpushedBranches(plan branch.CleanupPlan) []string {
	var skipBranches []string

	for _, ob := range plan.OrphanedBranches {
		if ob.HasUnpushed {
			commands.PrintWarningf("Branch '%s' in %s has %d unpushed commit(s)", ob.BranchName, ob.RepoName, ob.UnpushedCount)
			if !commands.PromptYesNo("Delete anyway? (y/n): ") {
				commands.PrintInfof("Skipping branch '%s'", ob.BranchName)
				skipBranches = append(skipBranches, fmt.Sprintf("%s:%s", ob.RepoName, ob.BranchName))
			}
		}
	}

	return skipBranches
}

func displayCleanupResult(result branch.CleanupResult) {
	fmt.Printf("\n")

	for _, d := range result.Deleted {
		commands.PrintSuccessf("Deleted %s in %s", d.BranchName, d.RepoName)
	}

	for _, f := range result.Failed {
		commands.PrintErrorf("Failed to delete %s in %s: %v", f.BranchName, f.RepoName, f.Error)
	}

	if len(result.Deleted) > 0 {
		commands.PrintSuccessf("Deleted %d branch(es)", len(result.Deleted))
	}
	if len(result.Failed) > 0 {
		commands.PrintWarningf("Failed to delete %d branch(es)", len(result.Failed))
	}
}

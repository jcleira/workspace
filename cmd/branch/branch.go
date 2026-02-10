// Package branch provides the workspace branch management commands.
package branch

import (
	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
)

var BranchCmd = &cobra.Command{
	Use:     "branch",
	Aliases: []string{"b"},
	Short:   "Manage branches in main repositories",
	Long:    `Manage branches in main repositories, including listing and cleaning up orphaned branches.`,
}

func init() {
	cmd.RootCmd.AddCommand(BranchCmd)
}

// Package config provides the workspace config command and subcommands.
package config

import (
	"github.com/spf13/cobra"

	"github.com/jcleira/workspace/cmd"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage workspace configuration",
	Long:  `Configure workspace settings like default directories and behavior.`,
}

func init() {
	cmd.RootCmd.AddCommand(ConfigCmd)
}

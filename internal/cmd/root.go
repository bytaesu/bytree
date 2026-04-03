package cmd

import (
	"github.com/spf13/cobra"
)

// NewRoot creates the root bytree command with all subcommands.
func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "bytree",
		Short: "Git worktree manager that copies excluded files",
		Long: `bytree creates isolated git worktrees and automatically copies files
matching patterns in .git/info/exclude from the main repository.

Useful for AI agents and multi-workspace setups where each worktree
needs local configs (.claude/, IDE settings, etc.) that aren't tracked by git.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	root.AddCommand(
		newAddCmd(),
		newRemoveCmd(),
		newListCmd(),
		newSyncCmd(),
		newExcludedCmd(),
	)

	return root
}

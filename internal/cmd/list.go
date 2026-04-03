package cmd

import (
	"fmt"

	"github.com/bytaesu/bytree/internal/git"
	"github.com/bytaesu/bytree/internal/ui"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all bytree worktrees",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList()
		},
	}
}

func runList() error {
	repoRoot, err := git.RepoRoot(".")
	if err != nil {
		return err
	}

	repoName, err := git.RepoName(repoRoot)
	if err != nil {
		return err
	}

	worktrees, err := git.ListWorktrees(repoRoot)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(ui.Title.Render("bytree list"))
	ui.Dimf("%s", repoName)
	fmt.Println()

	if len(worktrees) == 0 {
		ui.Dimf("No worktrees found.")
		ui.Dimf("Create one: bytree add <name>")
		return nil
	}

	for _, wt := range worktrees {
		fmt.Println(ui.Bold.Render(wt.Name))
		ui.Dimf("  %s", wt.Path)
		ui.Dimf("  %s", wt.Branch)
		fmt.Println()
	}

	return nil
}

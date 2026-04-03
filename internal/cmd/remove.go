package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/bytaesu/bytree/internal/git"
	"github.com/bytaesu/bytree/internal/ui"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a worktree and its branch",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(args[0])
		},
	}
}

func runRemove(name string) error {
	repoRoot, err := git.RepoRoot(".")
	if err != nil {
		return err
	}

	repoName, err := git.RepoName(repoRoot)
	if err != nil {
		return err
	}

	worktreeBase := filepath.Join(repoRoot, "..", repoName+"-bytree")
	wtPath := filepath.Join(worktreeBase, name)
	branch := git.BranchPrefix + name

	fmt.Println()
	fmt.Println(ui.Title.Render("bytree remove"))
	fmt.Println()

	if err := git.RemoveWorktree(repoRoot, wtPath, branch); err != nil {
		return err
	}
	ui.Successf("Removed %s", name)

	return nil
}

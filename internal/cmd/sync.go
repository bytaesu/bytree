package cmd

import (
	"fmt"

	"github.com/bytaesu/bytree/internal/git"
	"github.com/bytaesu/bytree/internal/ui"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Copy excluded files into the current worktree",
		Long: `Copies files matching .git/info/exclude patterns from the main
repository into the current worktree. Run this from inside a worktree
to refresh local configs without recreating the worktree.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync()
		},
	}
}

func runSync() error {
	repoRoot, err := git.RepoRoot(".")
	if err != nil {
		return err
	}

	toplevel, err := git.WorktreeRoot(".")
	if err != nil {
		return err
	}

	if toplevel == repoRoot {
		return fmt.Errorf("not inside a worktree (this is the main repository)")
	}

	fmt.Println()
	fmt.Println(ui.Title.Render("bytree sync"))
	fmt.Println()

	var copied []string
	err = ui.SpinWhile("Copying excluded files...", func() error {
		var e error
		copied, e = git.CopyExcludedFiles(repoRoot, toplevel)
		return e
	})
	if err != nil {
		return err
	}

	if len(copied) > 0 {
		ui.Successf("Synced %d excluded file(s)", len(copied))
	} else {
		ui.Dimf("  No excluded files to copy")
	}

	return nil
}

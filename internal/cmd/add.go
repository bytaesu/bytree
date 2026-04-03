package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/bytaesu/bytree/internal/git"
	"github.com/bytaesu/bytree/internal/ui"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var baseBranch string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a worktree with excluded files copied",
		Args:  cobra.ExactArgs(1),
		Example: `  bytree add feature-x
  bytree add issue-123 --base develop`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(args[0], baseBranch)
		},
	}

	cmd.Flags().StringVar(&baseBranch, "base", "", "base branch (default: auto-detect)")
	return cmd
}

func runAdd(name, baseBranch string) error {
	repoRoot, err := git.RepoRoot(".")
	if err != nil {
		return err
	}

	repoName, err := git.RepoName(repoRoot)
	if err != nil {
		return err
	}

	if baseBranch == "" {
		baseBranch = git.DefaultBranch(repoRoot)
	}

	worktreeBase := filepath.Join(repoRoot, "..", repoName+"-bytree")

	fmt.Println()
	fmt.Println(ui.Title.Render("bytree add"))
	ui.Dimf("%s | base: %s", repoName, baseBranch)
	fmt.Println()

	var wt *git.Worktree
	err = ui.SpinWhile("Creating worktree...", func() error {
		var e error
		wt, e = git.CreateWorktree(repoRoot, worktreeBase, name, baseBranch)
		return e
	})
	if err != nil {
		return err
	}
	ui.Successf("Created %s", wt.Path)
	ui.Dimf("  Branch: %s", wt.Branch)
	fmt.Println()

	var copied []string
	err = ui.SpinWhile("Copying excluded files...", func() error {
		var e error
		copied, e = git.CopyExcludedFiles(repoRoot, wt.Path)
		return e
	})
	if err != nil {
		return err
	}

	if len(copied) > 0 {
		ui.Successf("Copied %d excluded file(s)", len(copied))
	} else {
		ui.Dimf("  No excluded files to copy")
	}

	fmt.Println()
	fmt.Println(ui.Bold.Render("Next:"))
	ui.Dimf("  cd %s", wt.Path)

	return nil
}

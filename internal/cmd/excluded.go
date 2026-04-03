package cmd

import (
	"fmt"

	"github.com/bytaesu/bytree/internal/git"
	"github.com/bytaesu/bytree/internal/ui"
	"github.com/spf13/cobra"
)

func newExcludedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "excluded",
		Short: "Show patterns in .git/info/exclude",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExcluded()
		},
	}
}

func runExcluded() error {
	repoRoot, err := git.RepoRoot(".")
	if err != nil {
		return err
	}

	patterns, err := git.ExcludePatterns(repoRoot)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(ui.Title.Render("bytree excluded"))
	fmt.Println()

	if len(patterns) == 0 {
		ui.Dimf("No patterns in .git/info/exclude")
		return nil
	}

	for _, p := range patterns {
		fmt.Println(p)
	}

	return nil
}

package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// RepoRoot returns the root of the main repository (not a worktree).
func RepoRoot(cwd string) (string, error) {
	commonDir, err := run(cwd, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	if commonDir == ".git" {
		return run(cwd, "rev-parse", "--show-toplevel")
	}

	// commonDir is an absolute path like /path/to/repo/.git
	return filepath.Dir(commonDir), nil
}

// RepoName returns the repository name parsed from the origin remote URL.
// Falls back to the directory name if no origin remote is configured.
func RepoName(repoRoot string) (string, error) {
	url, err := run(repoRoot, "remote", "get-url", "origin")
	if err != nil {
		return filepath.Base(repoRoot), nil
	}

	re := regexp.MustCompile(`[/:]([^/]+?)(?:\.git)?$`)
	m := re.FindStringSubmatch(url)
	if m == nil {
		return filepath.Base(repoRoot), nil
	}
	return m[1], nil
}

// DefaultBranch detects the default branch (main, master, etc.).
func DefaultBranch(repoRoot string) string {
	ref, err := run(repoRoot, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		return strings.TrimPrefix(ref, "refs/remotes/origin/")
	}

	branches, err := run(repoRoot, "branch", "-r")
	if err == nil {
		if strings.Contains(branches, "origin/main") {
			return "main"
		}
		if strings.Contains(branches, "origin/master") {
			return "master"
		}
	}
	return "main"
}

// WorktreeRoot returns the toplevel of the current working tree.
func WorktreeRoot(cwd string) (string, error) {
	return run(cwd, "rev-parse", "--show-toplevel")
}

// run executes a git command and returns trimmed stdout.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// runNoFail executes a git command, ignoring errors.
func runNoFail(dir string, args ...string) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	_ = cmd.Run()
}

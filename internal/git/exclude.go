package git

import (
	"os"
	"path/filepath"
	"strings"
)

// ExcludePatterns reads patterns from .git/info/exclude.
func ExcludePatterns(repoRoot string) ([]string, error) {
	excludePath := filepath.Join(repoRoot, ".git", "info", "exclude")
	data, err := os.ReadFile(excludePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, nil
}

// ExcludeEntries returns top-level entries (files or directories) that match
// exclude patterns using git ls-files. The --directory flag collapses directories
// into single entries, avoiding expensive recursive walks.
func ExcludeEntries(repoRoot string) ([]string, error) {
	excludeFile := filepath.Join(repoRoot, ".git", "info", "exclude")
	if _, err := os.Stat(excludeFile); os.IsNotExist(err) {
		return nil, nil
	}

	out, err := run(repoRoot, "ls-files", "-o", "-i",
		"--exclude-from="+excludeFile, "--directory")
	if err != nil {
		return nil, err
	}

	if out == "" {
		return nil, nil
	}

	seen := make(map[string]struct{})
	var entries []string

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// git ls-files --directory appends / to directories
		line = strings.TrimSuffix(line, "/")

		if _, ok := seen[line]; !ok {
			seen[line] = struct{}{}
			entries = append(entries, line)
		}
	}

	return entries, nil
}

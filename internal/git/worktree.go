package git

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const BranchPrefix = "bytree/"

// Buffer pool to reduce GC pressure during file copies.
var bufPool = sync.Pool{
	New: func() any { return make([]byte, 32*1024) },
}

// Worktree represents a bytree-managed worktree.
type Worktree struct {
	Name   string
	Path   string
	Branch string
}

// CreateWorktree creates a new git worktree with a bytree/* branch.
func CreateWorktree(repoRoot, worktreeBase, name, baseBranch string) (*Worktree, error) {
	branch := BranchPrefix + name
	wtPath := filepath.Join(worktreeBase, name)

	if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	// Clean up any existing worktree/branch
	if _, err := os.Stat(wtPath); err == nil {
		runNoFail(repoRoot, "worktree", "remove", wtPath, "--force")
		runNoFail(repoRoot, "branch", "-D", branch)
	}

	if _, err := run(repoRoot, "worktree", "add", "-b", branch, wtPath, baseBranch); err != nil {
		return nil, fmt.Errorf("git worktree add: %w", err)
	}

	return &Worktree{Name: name, Path: wtPath, Branch: branch}, nil
}

// RemoveWorktree removes a worktree and deletes its branch.
func RemoveWorktree(repoRoot, wtPath, branch string) error {
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worktree %q not found", wtPath)
	}
	runNoFail(repoRoot, "worktree", "remove", wtPath, "--force")
	runNoFail(repoRoot, "branch", "-D", branch)
	return nil
}

// ListWorktrees returns all bytree-managed worktrees.
func ListWorktrees(repoRoot string) ([]Worktree, error) {
	out, err := run(repoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var (
		result      []Worktree
		currentPath string
	)
	for line := range strings.SplitSeq(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			currentPath = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch refs/heads/"+BranchPrefix):
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			name := strings.TrimPrefix(branch, BranchPrefix)
			result = append(result, Worktree{
				Name:   name,
				Path:   currentPath,
				Branch: branch,
			})
		}
	}
	return result, nil
}

// CopyExcludedFiles copies excluded entries from repoRoot into the worktree.
// Directories are walked and all files are copied concurrently.
func CopyExcludedFiles(repoRoot, wtPath string) ([]string, error) {
	entries, err := ExcludeEntries(repoRoot)
	if err != nil {
		return nil, err
	}

	// Collect all individual files to copy.
	type job struct{ src, dst, rel string }
	type dirMtime struct {
		path string
		t    time.Time
	}
	var jobs []job
	var dirMtimes []dirMtime

	for _, rel := range entries {
		src := filepath.Join(repoRoot, rel)
		dst := filepath.Join(wtPath, rel)

		srcInfo, err := os.Lstat(src)
		if err != nil {
			continue
		}

		if srcInfo.IsDir() {
			// If the destination directory exists and has the same mtime,
			// the source hasn't changed since last sync. Skip the entire tree.
			if dstInfo, err := os.Lstat(dst); err == nil && dstInfo.IsDir() {
				if srcInfo.ModTime().Equal(dstInfo.ModTime()) {
					continue
				}
			}
			_ = filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				r, _ := filepath.Rel(repoRoot, path)
				jobs = append(jobs, job{
					src: path,
					dst: filepath.Join(wtPath, r),
					rel: r,
				})
				return nil
			})
			dirMtimes = append(dirMtimes, dirMtime{dst, srcInfo.ModTime()})
		} else {
			jobs = append(jobs, job{src: src, dst: dst, rel: rel})
		}
	}

	// Pre-create all needed directories in a single pass.
	dirs := make(map[string]struct{})
	for _, j := range jobs {
		dirs[filepath.Dir(j.dst)] = struct{}{}
	}
	for d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir: %w", err)
		}
	}

	// Copy files concurrently with bounded parallelism.
	const workers = 16
	sem := make(chan struct{}, workers)
	var (
		mu      sync.Mutex
		copyErr error
	)
	var wg sync.WaitGroup

	for i := range jobs {
		if copyErr != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(j job) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := copyFile(j.src, j.dst); err != nil {
				mu.Lock()
				if copyErr == nil {
					copyErr = fmt.Errorf("copy %s: %w", j.rel, err)
				}
				mu.Unlock()
			}
		}(jobs[i])
	}
	wg.Wait()

	if copyErr != nil {
		return nil, copyErr
	}

	// Preserve directory mtimes after all copies are done.
	for _, dm := range dirMtimes {
		_ = os.Chtimes(dm.path, dm.t, dm.t)
	}

	// Return entry-level names (not individual files).
	return entries, nil
}

func copyFile(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Skip if destination has same size and modification time.
	if dstInfo, err := os.Lstat(dst); err == nil {
		if info.Size() == dstInfo.Size() && info.ModTime().Equal(dstInfo.ModTime()) {
			return nil
		}
	}

	// Handle symlinks
	if info.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		_ = os.Remove(dst)
		return os.Symlink(target, dst)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}

	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	if _, err = io.CopyBuffer(out, in, buf); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	// Preserve modification time for future skip checks.
	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}

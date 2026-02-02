import { existsSync } from "node:fs";
import { readFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { $ } from "bun";

/**
 * Get the root of the main git repository (not worktree)
 */
export async function getRepoRoot(cwd: string): Promise<string> {
	const result = await $`git -C ${cwd} rev-parse --git-common-dir`.quiet();
	const gitCommonDir = result.text().trim();

	// If .git, we're in the main repo - use show-toplevel
	if (gitCommonDir === ".git") {
		const topLevel = await $`git -C ${cwd} rev-parse --show-toplevel`.quiet();
		return topLevel.text().trim();
	}

	// Otherwise, gitCommonDir is absolute path to main repo's .git
	// e.g., /path/to/repo/.git -> /path/to/repo
	return dirname(resolve(cwd, gitCommonDir));
}

/**
 * Get repository info from remote URL
 */
export async function getRepoInfo(
	repoRoot: string,
): Promise<{ owner: string; repo: string }> {
	const result = await $`git -C ${repoRoot} remote get-url origin`.quiet();
	const url = result.text().trim();

	// Parse GitHub URL (supports both HTTPS and SSH)
	const match = url.match(/github\.com[:/]([^/]+)\/([^/.]+)/);
	if (!match || !match[1] || !match[2]) {
		throw new Error(`Could not parse GitHub URL: ${url}`);
	}

	return { owner: match[1], repo: match[2] };
}

/**
 * Get the default branch name
 */
export async function getDefaultBranch(repoRoot: string): Promise<string> {
	try {
		const result =
			await $`git -C ${repoRoot} symbolic-ref refs/remotes/origin/HEAD`.quiet();
		const ref = result.text().trim();
		return ref.replace("refs/remotes/origin/", "");
	} catch {
		// Fallback to main or master
		const branches = await $`git -C ${repoRoot} branch -r`.quiet();
		const branchList = branches.text();
		if (branchList.includes("origin/main")) return "main";
		if (branchList.includes("origin/master")) return "master";
		return "main";
	}
}

/**
 * Get the current branch name
 */
export async function getCurrentBranch(repoRoot: string): Promise<string> {
	const result = await $`git -C ${repoRoot} branch --show-current`.quiet();
	return result.text().trim();
}

/**
 * Get the actual .git directory (handles worktrees where .git is a file)
 */
async function getGitDir(repoRoot: string): Promise<string> {
	const gitPath = join(repoRoot, ".git");

	// Use git to get the actual git directory
	try {
		const result =
			await $`git -C ${repoRoot} rev-parse --git-common-dir`.quiet();
		return result.text().trim();
	} catch {
		return gitPath;
	}
}

/**
 * Parse .git/info/exclude and return list of excluded patterns
 */
export async function getExcludePatterns(repoRoot: string): Promise<string[]> {
	const gitDir = await getGitDir(repoRoot);
	const excludePath = join(gitDir, "info", "exclude");
	if (!existsSync(excludePath)) {
		return [];
	}

	const content = await readFile(excludePath, "utf-8");
	const patterns: string[] = [];

	for (const line of content.split("\n")) {
		const trimmed = line.trim();
		// Skip comments and empty lines
		if (!trimmed || trimmed.startsWith("#")) {
			continue;
		}
		patterns.push(trimmed);
	}

	return patterns;
}

/**
 * Find files matching exclude patterns (files that exist but are ignored)
 */
export async function findExcludedFiles(repoRoot: string): Promise<string[]> {
	const patterns = await getExcludePatterns(repoRoot);
	if (patterns.length === 0) {
		return [];
	}

	const excludedFiles: string[] = [];

	for (let pattern of patterns) {
		// Remove leading slash for glob matching
		if (pattern.startsWith("/")) {
			pattern = pattern.slice(1);
		}

		// Handle directory patterns (ending with /)
		if (pattern.endsWith("/")) {
			pattern = `${pattern}**/*`;
		}

		try {
			const glob = new Bun.Glob(pattern);
			for await (const file of glob.scan({ cwd: repoRoot, dot: true })) {
				const fullPath = join(repoRoot, file);
				if (existsSync(fullPath)) {
					excludedFiles.push(file);
				}
			}
		} catch {
			// Skip invalid patterns
		}
	}

	return [...new Set(excludedFiles)]; // Remove duplicates
}

/**
 * Create a worktree with a unique branch name
 */
export async function createWorktree(
	repoRoot: string,
	worktreeBase: string,
	name: string,
	baseBranch: string,
): Promise<{ path: string; branch: string }> {
	const branch = `bytree/${name}`;
	const worktreePath = join(worktreeBase, name);

	// Create worktree directory if it doesn't exist
	await $`mkdir -p ${dirname(worktreePath)}`.quiet();

	// Remove existing worktree if it exists
	if (existsSync(worktreePath)) {
		await $`git -C ${repoRoot} worktree remove ${worktreePath} --force`
			.quiet()
			.nothrow();
		await $`git -C ${repoRoot} branch -D ${branch}`.quiet().nothrow();
	}

	// Create new worktree
	await $`git -C ${repoRoot} worktree add -b ${branch} ${worktreePath} ${baseBranch}`.quiet();

	return { path: worktreePath, branch };
}

/**
 * Copy excluded files to worktree
 */
export async function copyExcludedFiles(
	repoRoot: string,
	worktreePath: string,
): Promise<string[]> {
	const excludedFiles = await findExcludedFiles(repoRoot);
	const copied: string[] = [];

	for (const file of excludedFiles) {
		const srcPath = join(repoRoot, file);
		const destPath = join(worktreePath, file);

		// Create parent directory if needed
		await $`mkdir -p ${dirname(destPath)}`.quiet();

		// Copy file
		await $`cp -r ${srcPath} ${destPath}`.quiet();
		copied.push(file);
	}

	return copied;
}

/**
 * List all bytree worktrees
 */
export async function listMyWorktrees(
	repoRoot: string,
): Promise<Array<{ path: string; branch: string }>> {
	const result = await $`git -C ${repoRoot} worktree list --porcelain`.quiet();
	const output = result.text();

	const worktrees: Array<{ path: string; branch: string }> = [];
	let currentPath = "";
	let currentBranch = "";

	for (const line of output.split("\n")) {
		if (line.startsWith("worktree ")) {
			currentPath = line.replace("worktree ", "");
		} else if (line.startsWith("branch ")) {
			currentBranch = line.replace("branch refs/heads/", "");
			// Only include "bytree/" branches
			if (currentBranch.startsWith("bytree/")) {
				worktrees.push({ path: currentPath, branch: currentBranch });
			}
		}
	}

	return worktrees;
}

/**
 * Remove a worktree and its branch
 */
export async function removeWorktree(
	repoRoot: string,
	worktreePath: string,
	branch: string,
): Promise<void> {
	await $`git -C ${repoRoot} worktree remove ${worktreePath} --force`
		.quiet()
		.nothrow();
	await $`git -C ${repoRoot} branch -D ${branch}`.quiet().nothrow();
}

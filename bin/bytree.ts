#!/usr/bin/env bun

/**
 * bytree - Git worktree manager that copies excluded files
 */

import { join } from "node:path";
import chalk from "chalk";
import {
	copyExcludedFiles,
	createWorktree,
	getDefaultBranch,
	getExcludePatterns,
	getRepoInfo,
	getRepoRoot,
	listMyWorktrees,
	removeWorktree,
} from "../src/lib/git";

const HELP = `
${chalk.cyan.bold("bytree")} - Git worktree manager

${chalk.white("Usage:")}
  bytree add <name>       Create a worktree with excluded files copied
  bytree remove <name>    Remove a worktree
  bytree list             List all bytree worktrees
  bytree excluded         Show patterns in .git/info/exclude

${chalk.white("Examples:")}
  bytree add feature-x    Create worktree at ../<repo>-bytree/feature-x
  bytree add issue-123    Create worktree for issue #123
  bytree remove feature-x Remove the worktree

${chalk.white("Options:")}
  --base <branch>    Base branch (default: auto-detect)
  --help, -h         Show help
  --version, -v      Show version
`;

async function addCommand(name: string, baseBranch?: string): Promise<void> {
	const cwd = process.cwd();
	const repoRoot = await getRepoRoot(cwd);
	const repoInfo = await getRepoInfo(repoRoot);
	const base = baseBranch || (await getDefaultBranch(repoRoot));
	const worktreeBase = join(repoRoot, "..", `${repoInfo.repo}-bytree`);

	console.log();
	console.log(chalk.cyan.bold("bytree add"));
	console.log(chalk.gray(`${repoInfo.owner}/${repoInfo.repo} | base: ${base}`));
	console.log();

	// Create worktree
	console.log(chalk.gray("Creating worktree..."));
	const { path: worktreePath, branch } = await createWorktree(
		repoRoot,
		worktreeBase,
		name,
		base,
	);
	console.log(chalk.green(`✓ Created ${worktreePath}`));
	console.log(chalk.gray(`  Branch: ${branch}`));

	// Copy excluded files
	console.log();
	console.log(chalk.gray("Copying excluded files..."));
	const copied = await copyExcludedFiles(repoRoot, worktreePath);

	if (copied.length > 0) {
		console.log(chalk.green(`✓ Copied ${copied.length} excluded file(s)`));
	} else {
		console.log(chalk.gray("  No excluded files to copy"));
	}

	console.log();
	console.log(chalk.white("Next:"));
	console.log(chalk.gray(`  cd ${worktreePath}`));
}

async function removeCommand(name: string): Promise<void> {
	const cwd = process.cwd();
	const repoRoot = await getRepoRoot(cwd);
	const repoInfo = await getRepoInfo(repoRoot);
	const worktreeBase = join(repoRoot, "..", `${repoInfo.repo}-bytree`);
	const worktreePath = join(worktreeBase, name);
	const branch = `bytree/${name}`;

	console.log();
	console.log(chalk.cyan.bold("bytree remove"));
	console.log();

	await removeWorktree(repoRoot, worktreePath, branch);
	console.log(chalk.green(`✓ Removed ${name}`));
}

async function listCommand(): Promise<void> {
	const cwd = process.cwd();
	const repoRoot = await getRepoRoot(cwd);
	const repoInfo = await getRepoInfo(repoRoot);
	const worktrees = await listMyWorktrees(repoRoot);

	console.log();
	console.log(chalk.cyan.bold("bytree list"));
	console.log(chalk.gray(`${repoInfo.owner}/${repoInfo.repo}`));
	console.log();

	if (worktrees.length === 0) {
		console.log(chalk.gray("No worktrees found."));
		console.log(chalk.gray("Create one: bytree add <name>"));
		return;
	}

	for (const wt of worktrees) {
		const name = wt.branch.replace("bytree/", "");
		console.log(chalk.white(name));
		console.log(chalk.gray(`  ${wt.path}`));
		console.log(chalk.gray(`  ${wt.branch}`));
		console.log();
	}
}

async function excludedCommand(): Promise<void> {
	const cwd = process.cwd();
	const repoRoot = await getRepoRoot(cwd);

	console.log();
	console.log(chalk.cyan.bold("bytree excluded"));
	console.log();

	const patterns = await getExcludePatterns(repoRoot);
	if (patterns.length === 0) {
		console.log(chalk.gray("No patterns in .git/info/exclude"));
		return;
	}

	for (const pattern of patterns) {
		console.log(pattern);
	}
}

async function main() {
	const args = Bun.argv.slice(2);

	if (args.includes("--version") || args.includes("-v")) {
		console.log("0.1.0");
		process.exit(0);
	}

	if (args.length === 0 || args.includes("--help") || args.includes("-h")) {
		console.log(HELP);
		process.exit(0);
	}

	const command = args[0];

	try {
		switch (command) {
			case "add": {
				const name = args[1];
				if (!name) {
					console.error(chalk.red("Error: Name required"));
					console.error(chalk.gray("Usage: bytree add <name>"));
					process.exit(1);
				}

				let baseBranch: string | undefined;
				const baseIdx = args.indexOf("--base");
				if (baseIdx !== -1) {
					baseBranch = args[baseIdx + 1];
				}

				await addCommand(name, baseBranch);
				break;
			}

			case "remove": {
				const name = args[1];
				if (!name) {
					console.error(chalk.red("Error: Name required"));
					console.error(chalk.gray("Usage: bytree remove <name>"));
					process.exit(1);
				}
				await removeCommand(name);
				break;
			}

			case "list": {
				await listCommand();
				break;
			}

			case "excluded": {
				await excludedCommand();
				break;
			}

			default:
				console.error(chalk.red(`Error: Unknown command: ${command}`));
				console.log(HELP);
				process.exit(1);
		}
	} catch (error) {
		console.error(chalk.red("Error:"), error);
		process.exit(1);
	}
}

main();

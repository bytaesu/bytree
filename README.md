# bytree

Git worktree manager that automatically copies files excluded via `.git/info/exclude`

## Why?

AI agents work best for you with personalized settings like project instructions, local configs, and custom workflows. When running multiple agents, working in the same workspace can cause conflicts.

Isolated worktrees solve this, but `git worktree` doesn't copy files in `.git/info/exclude` (like `.claude/`, IDE settings). You have to manually copy them every time.

**bytree** solves this by automatically copying excluded files when creating a worktree.

## Install

```bash
bun add -g bytree
```

## Usage

### Setup excluded files

Add patterns to `.git/info/exclude` in your repository:

```txt
/.claude/
/__local__/

# ...
```

### Create a worktree

```bash
bytree add feature-x
```

This will:

1. Create a worktree at `../<repo>-bytree/feature-x`
2. Create branch `bytree/feature-x`
3. Copy all files matching patterns in `.git/info/exclude`

### List worktrees

```bash
bytree list
```

### Remove a worktree

```bash
bytree remove feature-x
```

### View exclude patterns

```bash
bytree excluded
```

## Options

```
--base <branch>    Base branch (default: auto-detect)
--help, -h         Show help
--version, -v      Show version
```

## License

MIT

# bytree

`git worktree` + your excluded files.

## Why?

AI agents work best with personalized settings (`.claude/`, `.codex/`, IDE configs). When running multiple agents in isolated worktrees, `git worktree` doesn't copy these excluded files. **bytree** does.

## Install

| Method   | Command                                                       |
| -------- | ------------------------------------------------------------- |
| Go       | `go install github.com/bytaesu/bytree/cmd/bytree@latest`      |
| npm      | `npm install -g bytree`                                       |
| Homebrew | `brew install bytaesu/tap/bytree`                             |
| Binary   | [GitHub Releases](https://github.com/bytaesu/bytree/releases) |

## Usage

First, define local settings you want copied across worktrees in `.git/info/exclude`:

```
/.claude/
/.codex/
/.github/prompts/
```

bytree reads these patterns to decide what to copy.

- ### `bytree add <name>`

  Create a worktree with excluded files copied automatically.

  ```bash
  bytree add feature-x
  bytree add issue-123 --base develop
  ```

  Worktrees are created at `../<repo>-bytree/<name>` on branch `bytree/<name>`.

- ### `bytree sync`

  Re-sync excluded files into the current worktree.

  ```bash
  cd path/to/worktree
  bytree sync
  ```

- ### `bytree list`

  List all bytree-managed worktrees.

- ### `bytree remove <name>`

  Remove a worktree and its branch.

- ### `bytree excluded`

  Show patterns in `.git/info/exclude`.

## How it works

Reads `.git/info/exclude` (not `.gitignore`), finds matching files, copies them. That's it.

## License

MIT

---
description: Create git commits with user approval and no Claude attribution
model: sonnet
---

# Commit Changes

Create git commits for changes in the working tree. This includes changes from the current session and any pre-existing staged or unstaged modifications.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.
Run `rpi --help` to discover available commands and `rpi <command> --help` for detailed usage with examples.

## Process

### 1. Understand the changes

Use `rpi` to gather consolidated git context (status, diff, recent commits, sensitive files).

This returns branch, status (tracked/untracked/staged files), diff summary, and recent commits as JSON. Use this to get the full picture.

Review the conversation history to understand the intent behind changes. If there are staged files not discussed in the conversation, inspect them too — the user may have edited files manually.

### 2. Check for problems

Before planning commits, scan the changeset for issues:

- **Sensitive files**: Use `rpi` to check staged files for sensitive content (`.env`, credentials, secrets, API keys, or large binaries). Warn the user and exclude them unless explicitly told otherwise.
- **Nothing to commit**: If the working tree is clean (no staged, unstaged, or untracked changes), tell the user and stop — don't create an empty commit.

### 3. Plan your commit(s)

- Group related files into logical, focused commits — prefer smaller over monolithic
- Draft commit messages in imperative mood, focusing on the "why" not the "what"
- Match the repo's existing commit style from the recent commits in the git context. If there aren't any commit conventions detected, follow commitizen convention:
  - `feat: add user authentication`
  - `fix: resolve null pointer in data loader`
  - `refactor: extract validation into shared module`
  - `test: add coverage for edge cases in parser`
  - `docs: update setup instructions`
  - `chore: remove unused dependencies`

### 4. Present the plan

For each planned commit, show:
- The files to be staged
- The commit message

Then ask: "I plan to create [N] commit(s) with these changes. Shall I proceed?"

### 5. Execute

On confirmation:

1. Stage specific files with `git add <file>` (never `-A` or `.`)
2. Commit using a HEREDOC to handle special characters and multi-line messages safely:
   ```bash
   git commit -m "$(cat <<'EOF'
   feat: add user authentication
   EOF
   )"
   ```
3. Repeat for each planned commit
4. Show the result with `git log --oneline -n [number of commits]`

### 6. If a hook blocks the commit

Pre-commit hooks (linting, formatting, type checking) may reject the commit. When this happens:

- Read the error output to understand what failed
- Fix the issues (run formatters, resolve lint errors, etc.)
- Re-stage the affected files
- Create a **new** commit — never use `--amend`, because the failed commit didn't actually happen and amending would modify the previous unrelated commit


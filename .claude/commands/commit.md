---
description: Create git commits with user approval and no Claude attribution
model: sonnet
---

# Commit Changes

Create git commits for changes in the working tree. This includes changes from the current session and any pre-existing staged or unstaged modifications.

## Process

### 1. Understand the changes

Run these commands in parallel to get the full picture:

- `git status` — see all tracked, untracked, and staged files (never use `-uall`)
- `git diff` — unstaged changes
- `git diff --staged` — already-staged changes
- `git log --oneline -n 10` — recent commits, so you can match the repo's message style

Review the conversation history to understand the intent behind changes. If there are staged files not discussed in the conversation, inspect them too — the user may have edited files manually.

### 2. Check for problems

Before planning commits, scan the changeset for issues:

- **Sensitive files**: If you see `.env`, credentials, secrets, API keys, or large binaries in the changeset, warn the user and exclude them unless explicitly told otherwise.
- **Nothing to commit**: If the working tree is clean (no staged, unstaged, or untracked changes), tell the user and stop — don't create an empty commit.

### 3. Plan your commit(s)

- Group related files into logical, focused commits — prefer smaller over monolithic
- Draft commit messages in imperative mood, focusing on the "why" not the "what"
- Match the repo's existing commit style based on `git log` output. If there aren't any commit conventions detected, follow commitizen convention:
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

## Attribution

- NEVER add co-author information or Claude attribution
- No "Generated with Claude" or "Co-Authored-By" lines
- Write messages as if the user wrote them

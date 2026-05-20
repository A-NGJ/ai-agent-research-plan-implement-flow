---
name: rpi-commit
description: Create git commits for the current working tree with user approval, sensitive-content scan, and gitignore enforcement. Use when user says 'commit this', 'commit the changes', 'make a commit', or 'let's commit what we have'.
---

# Commit Changes

## Goal

Create git commits for changes in the working tree with user approval. This includes changes from the current session and any pre-existing modifications.

## Invariants

- Gather consolidated git context first: branch, status, diff summary, recent commits
- Scan staged files for sensitive content (.env, credentials, secrets, API keys) — warn and exclude unless explicitly told otherwise
- Scan candidate files against `.gitignore` rules — warn and exclude any matches (commonly: a file tracked before being added to .gitignore) unless explicitly told to keep it
- If the working tree is clean, inform the user and stop — no empty commits
- Group related files into logical, focused commits — prefer smaller over monolithic
- Draft commit messages in imperative mood, matching the repo's existing commit style; fall back to commitizen convention if no style detected
- Cap each commit message at 50 words total (subject + body); if a draft exceeds the cap, condense before presenting for approval
- Present the commit plan (files to stage + commit message for each) and ask for approval before executing
- Stage specific files with `git add <file>` — never use `-A` or `.`
- Use HEREDOC for commit messages to handle special characters safely
- After hook failure: read error output, fix the issue, re-stage, create a new commit — never use `--amend` (the failed commit didn't happen)

## Principles

- Review conversation history to understand intent behind changes
- Inspect manually-edited files (staged but not discussed) before committing
